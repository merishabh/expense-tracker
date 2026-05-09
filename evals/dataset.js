export const TEST_CASES = [
  {
    name: "foodie question with profile",
    question: "Am I a foodie?",
    profile: [
      { type: "life_context", content: "User spends around ₹1,91,524 per month on average" },
      { type: "pattern", content: "User's top spending categories are Food, Shopping, and Travel" },
      { type: "pattern", content: "Late night Swiggy orders on weekdays" },
    ],
    requiredSignals: ["food", "₹"],
    forbiddenSignals: ["generally speaking", "many people", "it depends"],
    minScore: 0.75,
  },
  {
    name: "planning question with profile",
    question: "What should I watch out for this month?",
    profile: [
      { type: "life_context", content: "User spends around ₹1,91,524 per month on average" },
      { type: "pattern", content: "User's top spending categories are Food, Shopping, and Travel" },
      { type: "pattern", content: "Late night Swiggy orders on weekdays" },
    ],
    requiredSignals: ["swiggy", "food"],
    forbiddenSignals: ["track your expenses", "create a budget", "generally"],
    minScore: 0.75,
  },
  {
    name: "no profile — should not hallucinate",
    question: "Am I a foodie?",
    profile: null,
    requiredSignals: [],
    forbiddenSignals: ["₹1,91", "₹191"],
    minScore: 0.75,
  },
  {
    name: "goal question — should reference profile goals",
    question: "How should I plan my savings?",
    profile: [
      { type: "life_context", content: "User spends around ₹1,91,524 per month on average" },
      { type: "pattern", content: "User's top spending categories are Food, Shopping, and Travel" },
      { type: "pattern", content: "Overspends in travel months" },
      { type: "goal", content: "Save ₹3L for Europe trip by June" },
    ],
    requiredSignals: ["europe", "₹3"],
    forbiddenSignals: ["set a goal", "what are you saving for", "generally"],
    minScore: 0.75,
  },
];
