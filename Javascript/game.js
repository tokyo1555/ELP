// game.js
const { createDeck, shuffleDeck, drawCard, cardToString } = require("./deck");

class PlayerState {
  constructor(id, totalScore = 0) {
    this.id = id;
    this.name = `Joueur ${id}`;
    this.totalScore = totalScore;

    this.numberCards = [];
    this.modifiers = [];
    this.actionsInFront = [];

    this.busted = false;
    this.frozen = false;
    this.stopped = false;

    this.hasSecondChance = false;
  }

  isActive() {
    return !this.busted && !this.frozen && !this.stopped;
  }

  getUniqueNumberValues() {
    return new Set(this.numberCards.map((c) => c.value));
  }

  hasDuplicateOnAdd(lastValue) {
    const values = this.numberCards.map((c) => c.value);
    const occurrences = values.filter((v) => v === lastValue).length;
    return occurrences > 1;
  }

  handToString() {
    const nums = this.numberCards.map((c) => cardToString(c)).join(" ");
    const mods = this.modifiers.map((c) => cardToString(c)).join(" ");
    const acts = this.actionsInFront.map((c) => cardToString(c)).join(" ");

    const parts = [];
    parts.push(nums ? `Nombres: ${nums}` : "Nombres: (aucun)");
    if (mods) parts.push(`Modifs: [${mods}]`);
    if (acts) parts.push(`Actions: {${acts}}`);
    return parts.join(" | ");
  }

  // Calcul pur: ne modifie pas totalScore
  computeRoundScore() {
    if (this.busted || this.frozen) return 0;

    let sum = this.numberCards.reduce((acc, c) => acc + c.value, 0);

    for (const m of this.modifiers) {
      switch (m.kind) {
        case "plus2":
          sum += 2;
          break;
        case "plus4":
          sum += 4;
          break;
        case "plus6":
          sum += 6;
          break;
        case "plus8":
          sum += 8;
          break;
        case "plus10":
          sum += 10;
          break;
        case "x2":
          sum *= 2;
          break;
      }
    }

    if (this.getUniqueNumberValues().size >= 7) sum += 15;

    return sum;
  }

  // Applique le score une seule fois
  finalizeRoundScore() {
    const s = this.computeRoundScore();
    this.totalScore += s;
    return s;
  }
}

class Flip7Round {
  constructor(numPlayers, playerScores = []) {
    this.numPlayers = numPlayers;
    this.players = Array.from({ length: numPlayers }, (_, i) => new PlayerState(i + 1, playerScores[i] || 0));
    this.deck = shuffleDeck(createDeck());
    this.roundOver = false;
  }

  dealInitialCards() {
    console.log("\nDistribution initiale :");
    for (const player of this.players) {
      if (this.roundOver) break;
      const card = drawCard(this.deck);
      if (!card) break;
      this.resolveDraw(player, card, { initialDeal: true });
      console.log(`- ${player.name} pioche ${cardToString(card)}`);
    }
  }

  drawForPlayer(player) {
    const card = drawCard(this.deck);
    if (!card) return null;
    this.resolveDraw(player, card);
    return card;
  }

  resolveDraw(player, card, { fromFlipThree = false } = {}) {
    if (this.roundOver) return;
    if (player.busted || player.frozen) return;

    if (card.type === "number") {
      player.numberCards.push(card);

      const value = card.value;
      if (player.hasDuplicateOnAdd(value)) {
        if (player.hasSecondChance) {
          player.hasSecondChance = false;
          player.actionsInFront = player.actionsInFront.filter((a) => a.kind !== "secondChance");
          player.numberCards.pop();
          console.log(`${player.name} utilise 2e chance: doublon évité (${value}).`);
        } else {
          player.busted = true;
          player.numberCards = [];
          console.log(`${player.name} fait un doublon (${value}): 0 point, éliminé.`);
        }
        return;
      }

      if (player.getUniqueNumberValues().size >= 7 && !fromFlipThree) {
        console.log(`${player.name} fait FLIP 7. Fin immédiate de la manche.`);
        this.roundOver = true;
      }

      return;
    }

    if (card.type === "modifier") {
      player.modifiers.push(card);
      console.log(`${player.name} reçoit un modificateur: ${cardToString(card)}`);
      return;
    }

    if (card.type === "action") {
      switch (card.kind) {
        case "freeze":
          player.frozen = true;
          player.numberCards = [];
          console.log(`${player.name} subit GEL: 0 point, éliminé.`);
          break;

        case "flipThree":
          console.log(`${player.name} joue TROIS: pioche 3 cartes.`);
          for (let i = 0; i < 3; i++) {
            if (this.roundOver || player.busted || player.frozen) break;
            const extra = drawCard(this.deck);
            if (!extra) break;
            console.log(`  ${i + 1}/3 -> ${cardToString(extra)}`);
            this.resolveDraw(player, extra, { fromFlipThree: true });
          }
          break;

        case "secondChance":
          if (!player.hasSecondChance) {
            player.hasSecondChance = true;
            player.actionsInFront.push(card);
            console.log(`${player.name} reçoit 2e chance: protège contre 1 doublon.`);
          } else {
            console.log(`${player.name} a déjà 2e chance: carte défaussée.`);
          }
          break;
      }
    }
  }

  isRoundOver() {
    return this.roundOver || this.players.every((p) => p.busted || p.frozen || p.stopped);
  }

  resetSecondChances() {
    this.players.forEach((p) => {
      p.hasSecondChance = false;
      p.actionsInFront = [];
    });
  }
}

module.exports = { Flip7Round };
