// deck.js
function createDeck() {
  const deck = [];

  // Cartes nombres : 1x0, 1x1, 2x2, ..., 12x12
  deck.push({ type: "number", value: 0 });
  for (let value = 1; value <= 12; value++) {
    for (let i = 0; i < value; i++) {
      deck.push({ type: "number", value });
    }
  }

  // Modificateurs
  const modifiers = [
    { kind: "plus2" },
    { kind: "plus2" },
    { kind: "plus4" },
    { kind: "plus4" },
    { kind: "plus6" },
    { kind: "plus6" },
    { kind: "plus8" },
    { kind: "plus10" },
    { kind: "x2" },
    { kind: "x2" },
  ];
  modifiers.forEach((m) => deck.push({ type: "modifier", ...m }));

  // Actions
  const actions = [
    { kind: "freeze" },
    { kind: "freeze" },
    { kind: "freeze" },
    { kind: "flipThree" },
    { kind: "flipThree" },
    { kind: "secondChance" },
    { kind: "secondChance" },
  ];
  actions.forEach((a) => deck.push({ type: "action", ...a }));

  return deck;
}

function shuffleDeck(deck) {
  for (let i = deck.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [deck[i], deck[j]] = [deck[j], deck[i]];
  }
  return deck;
}

function drawCard(deck) {
  return deck.length > 0 ? deck.shift() : null;
}

function cardToString(card) {
  if (card.type === "number") return String(card.value);

  if (card.type === "modifier") {
    switch (card.kind) {
      case "plus2":
        return "+2";
      case "plus4":
        return "+4";
      case "plus6":
        return "+6";
      case "plus8":
        return "+8";
      case "plus10":
        return "+10";
      case "x2":
        return "x2";
    }
  }

  if (card.type === "action") {
    switch (card.kind) {
      case "freeze":
        return "GEL";
      case "flipThree":
        return "TROIS";
      case "secondChance":
        return "2eCHANCE";
    }
  }

  return "?";
}

module.exports = { createDeck, shuffleDeck, drawCard, cardToString };
