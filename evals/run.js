import Anthropic from "@anthropic-ai/sdk";
import { TEST_CASES } from "./dataset.js";
import { checkSignals, judgeAnswer } from "./scorer.js";

const client = new Anthropic({ apiKey: process.env.ANTHROPIC_API_KEY });

// Mirrors LoadMemories() formatting from services/memory_service.go
function buildSystemPrompt(profile) {
  let prompt = "You are a finance agent.";
  if (!profile) return prompt;

  const memoryBlock = profile
    .map((m) => `- [${m.type}] ${m.content}`)
    .join("\n");
  return `${prompt}\n\nWhat you know about this user:\n${memoryBlock}`;
}

export async function callAgent(question, profile) {
  const systemPrompt = buildSystemPrompt(profile);

  const response = await client.messages.create({
    model: "claude-sonnet-4-6",
    max_tokens: 512,
    system: systemPrompt,
    messages: [{ role: "user", content: question }],
  });

  return response.content[0].text;
}

async function runEvals() {
  const results = [];

  for (const [index, tc] of TEST_CASES.entries()) {
    console.log(`\n=== [${index + 1}/${TEST_CASES.length}] ${tc.name} ===`);

    const answer = await callAgent(tc.question, tc.profile);
    const signals = checkSignals(answer, tc.requiredSignals, tc.forbiddenSignals);

    if (!signals.passed) {
      console.log(`❌ ${tc.name}`);
      if (signals.missingSignals.length > 0)
        console.log(`   Missing signals: ${signals.missingSignals.join(", ")}`);
      if (signals.triggeredForbidden.length > 0)
        console.log(`   Forbidden triggered: ${signals.triggeredForbidden.join(", ")}`);
      console.log(`   Answer preview: ${answer.slice(0, 100)}`);
      results.push({ name: tc.name, score: 0, passed: false });
      continue;
    }

    const judgment = await judgeAnswer(tc.question, answer, tc.profile);
    const passed = judgment.score >= tc.minScore;
    console.log(`${passed ? "✅" : "❌"} ${tc.name}`);
    console.log(`   Score: ${judgment.score} (min: ${tc.minScore})`);
    console.log(`   Reason: ${judgment.reason}`);
    results.push({ name: tc.name, score: judgment.score, passed });
  }

  const totalPassed = results.filter((r) => r.passed).length;
  const avgScore = results.reduce((sum, r) => sum + r.score, 0) / results.length;

  console.log("\n========== SUMMARY ==========");
  console.log(`Passed: ${totalPassed}/${results.length}`);
  console.log(`Average score: ${avgScore.toFixed(2)}`);
  console.log(
    avgScore >= 0.75
      ? "LangMem memory is solid ✅"
      : "It needs work ❌"
  );
}

runEvals();
