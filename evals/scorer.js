import Anthropic from "@anthropic-ai/sdk";

const judgeClient = new Anthropic({ apiKey: process.env.ANTHROPIC_API_KEY });

const JUDGE_SYSTEM_PROMPT = `You score personal finance agent answers on personalisation quality.
Return ONLY valid JSON, no other text: { "score": 0.0-1.0, "reason": "one sentence" }

Scoring guide:
1.0 — answer uses specific numbers, named merchants (Swiggy/Amazon etc),
      references this person's actual patterns. Feels like personal advice.
0.7 — somewhat specific but contains some generic elements
0.5 — half personalised, half generic
0.3 — mostly generic with minor specific elements
0.0 — completely generic advice OR hallucinated facts not in the profile`;

export function checkSignals(answer, requiredSignals, forbiddenSignals) {
  const lower = answer.toLowerCase();

  const missingSignals = requiredSignals.filter(
    (s) => !lower.includes(s.toLowerCase())
  );

  const triggeredForbidden = forbiddenSignals.filter((s) =>
    lower.includes(s.toLowerCase())
  );

  return {
    passed: missingSignals.length === 0 && triggeredForbidden.length === 0,
    missingSignals,
    triggeredForbidden,
  };
}

export async function judgeAnswer(question, answer, profile) {
  const profileText = profile
    ? profile.map((m) => `- [${m.type}] ${m.content}`).join("\n")
    : "No profile available.";

  const userMessage = `Question asked: ${question}

Profile available to the agent:
${profileText}

Agent's answer:
${answer}`;

  try {
    const response = await judgeClient.messages.create({
      model: "claude-sonnet-4-6",
      max_tokens: 256,
      system: JUDGE_SYSTEM_PROMPT,
      messages: [{ role: "user", content: userMessage }],
    });

    const raw = response.content[0].text.trim().replace(/^```[a-z]*\n?|```$/g, "").trim();
    const parsed = JSON.parse(raw);
    return { score: parsed.score, reason: parsed.reason };
  } catch {
    return { score: 0.5, reason: "judge parse failed" };
  }
}
