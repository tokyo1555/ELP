// logger.js
const fs = require("fs");

class GameLogger {
  constructor(filename = "games.json") {
    this.filename = filename;
    this.data = { games: [] };
    this.load();
  }

  load() {
    try {
      const content = fs.readFileSync(this.filename, "utf8");
      const parsed = JSON.parse(content);
      this.data = parsed && Array.isArray(parsed.games) ? parsed : { games: [] };
    } catch (err) {
      console.log("📁 Nouveau fichier games.json");
      this.data = { games: [] };
    }
  }

  saveRound(numPlayers, players) {
    const gameId = this.data.games.length + 1;

    const roundData = {
      id: gameId,
      date: new Date().toISOString(),
      numPlayers,
      players: players.map((p) => ({
        name: p.name,
        numberCards: p.numberCards.map((c) => ({ type: "number", value: c.value })),
        modifiers: p.modifiers.map((m) => ({ type: "modifier", kind: m.kind })),
        actionsInFront: p.actionsInFront.map((a) => ({ type: "action", kind: a.kind })),
        busted: p.busted,
        frozen: p.frozen,
        stopped: p.stopped,
        roundScore: p.computeRoundScore(), // pure
        totalScore: p.totalScore,
      })),
    };

    this.data.games.push(roundData);
    fs.writeFileSync(this.filename, JSON.stringify(this.data, null, 2), "utf8");
    console.log(`📝 Manche ${gameId} sauvegardée (${this.filename})`);
  }
}

module.exports = GameLogger;
