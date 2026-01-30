// index.js
const readline = require("readline");
const fs = require("fs");
const { Flip7Round } = require("./game");
const GameLogger = require("./logger");
const { cardToString } = require("./deck");

function askQuestion(rl, text) {
  return new Promise((resolve) => rl.question(text, resolve));
}

function normalizeYesNo(input) {
  const s = String(input || "").trim().toLowerCase();
  if (["o", "oui", "y", "yes"].includes(s)) return true;
  if (["n", "non", "no"].includes(s)) return false;
  return null;
}

function clampInt(n, min, max, fallback) {
  const v = Number.parseInt(n, 10);
  if (Number.isNaN(v)) return fallback;
  return Math.min(max, Math.max(min, v));
}

function printDivider(title = "") {
  const line = "=".repeat(44);
  if (!title) {
    console.log(line);
    return;
  }
  const t = ` ${title} `;
  const left = Math.max(0, Math.floor((44 - t.length) / 2));
  const right = Math.max(0, 44 - t.length - left);
  console.log("=".repeat(left) + t + "=".repeat(right));
}

function printScoreboard(players) {
  console.log("\nScore total :");
  players.forEach((p) => {
    const flags = [];
    if (p.busted) flags.push("ELIMINÉ");
    if (p.frozen) flags.push("GEL");
    if (p.stopped) flags.push("STOP");
    const status = flags.length ? ` [${flags.join(", ")}]` : "";
    console.log(`- ${p.name.padEnd(10)} : ${String(p.totalScore).padStart(3)} pts${status}`);
  });
}

async function main() {
  const rl = readline.createInterface({ input: process.stdin, output: process.stdout });

  const logger = new GameLogger();
  logger.data = { games: [] };
  fs.writeFileSync(logger.filename, JSON.stringify({ games: [] }, null, 2), "utf8");

  printDivider("FLIP 7");
  console.log("Nouvelle partie (historique réinitialisé : games.json)\n");

  const nbStr = await askQuestion(rl, "Nombre de joueurs (min 2, max 8) : ");
  const numPlayers = clampInt(nbStr, 2, 8, 2);

  const playerScores = Array(numPlayers).fill(0);
  console.log(`\nPartie à ${numPlayers} joueurs. Objectif : 200 points.\n`);

  let manche = 1;

  while (true) {
    printDivider(`MANCHE ${manche}`);
    const round = new Flip7Round(numPlayers, playerScores);
    round.dealInitialCards();

    let currentIndex = 0;

    while (!round.isRoundOver()) {
      const player = round.players[currentIndex];

      if (player.isActive()) {
        console.log("\n" + "-".repeat(44));
        console.log(`${player.name} | Score total du joueur : ${player.totalScore} pts`);
        console.log(`Main du joueur : ${player.handToString()}`);

        let decision = null;
        while (decision === null) {
          const answer = await askQuestion(rl, "Continuer ? (o/n) : ");
          decision = normalizeYesNo(answer);
        }

        if (!decision) {
          player.stopped = true;
          console.log(`${player.name} s'arrête.`);
        } else {
          const card = round.drawForPlayer(player);
          if (!card) {
            console.log("Plus de cartes. Fin.");
            player.stopped = true;
          } else {
            console.log(`${player.name} pioche : ${cardToString(card)}`);
            console.log(`Main du joueur : ${player.handToString()}`);
          }
        }
      }

      currentIndex = (currentIndex + 1) % numPlayers;
    }

    // Fin de manche: appliquer les scores UNE seule fois
    printDivider("FIN DE MANCHE");

    round.resetSecondChances();

    const roundScores = round.players.map((p) => p.finalizeRoundScore());
    round.players.forEach((p, i) => {
      console.log(
        `${p.name.padEnd(10)} : +${String(roundScores[i]).padStart(3)} pts | total: ${p.totalScore}`
      );
    });

    // Mettre à jour les scores persistants pour la manche suivante
    for (let i = 0; i < numPlayers; i++) {
      playerScores[i] = round.players[i].totalScore;
    }

    logger.saveRound(numPlayers, round.players);

    // Vérifier gagnant
    const winner = round.players.find((p) => p.totalScore >= 200);
    if (winner) {
      printDivider("VICTOIRE");
      console.log(`${winner.name} gagne avec ${winner.totalScore} points.`);
      printScoreboard(round.players);
      break;
    }

    printScoreboard(round.players);
    console.log("\nNouvelle manche...\n");
    manche += 1;
  }

  rl.close();
}

main().catch(console.error);