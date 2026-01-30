const net = require("net");
const readline = require("readline");

const input = process.argv[2];
if (!input) {
  console.log("Usage: node client.js <IP>[:PORT]");
  process.exit(1);
}

const [host, portStr] = input.split(":");
const port = portStr ? Number(portStr) : 5000;

const rl = readline.createInterface({ input: process.stdin, output: process.stdout });

function ask(question) {
  return new Promise((resolve) => rl.question(question, resolve));
}

let myId = null;
let myName = null;
let inNamePrompt = true;
let buffer = "";
let pendingLines = []; // messages reçus pendant la saisie du nom

function prettyPrint(line) {
  // Filtrage + formatage des messages du serveur
  if (line.startsWith("WELCOME ")) {
    myId = line.slice("WELCOME ".length).trim(); // ex: "J1"
    console.log(`Connecté au serveur. Identifiant: ${myId}`);
    return;
  }

  if (line.startsWith("CMD:")) {
    // on masque les instructions protocole côté client
    return;
  }

  if (line.startsWith("Puis:")) {
    // on masque
    return;
  }

  if (line.startsWith("LOBBY ")) {
    // ex: "LOBBY 1 joueurs | READY 0/1"
    console.log(`Salon: ${line.replace("LOBBY ", "")}`);
    return;
  }

  if (line.startsWith("JOUEURS ")) {
    // ex: "JOUEURS Anas:READY | ... "
    console.log(`Joueurs: ${line.replace("JOUEURS ", "")}`);
    return;
  }

  if (line.startsWith("INFO ")) {
    console.log(line.replace("INFO ", "• "));
    return;
  }

  if (line.startsWith("ERR ")) {
    console.log(`Erreur: ${line.replace("ERR ", "")}`);
    return;
  }

  if (line === "GAME_START") {
    console.log("Partie: démarrage");
    return;
  }

  if (line.startsWith("NB_JOUEURS ")) {
    console.log(`Nombre de joueurs: ${line.replace("NB_JOUEURS ", "")}`);
    return;
  }

  // fallback minimal (au cas où)
  console.log(line);
}

function handleServerLine(line) {
  if (!line) return;

  // Tant que le joueur saisit son nom, on stocke au lieu d'afficher (évite le bazar)
  if (inNamePrompt) {
    pendingLines.push(line);
    return;
  }

  prettyPrint(line);
}

async function main() {
  const socket = net.createConnection({ host, port }, async () => {
    console.log(`Connexion à ${host}:${port}...`);

    // Saisie nom (propre)
    const name = (await ask("Entre ton nom: ")).trim();
    myName = name || "Joueur";
    socket.write(`NAME ${myName}\n`);
    inNamePrompt = false;

    // Afficher les messages reçus pendant la saisie, proprement
    if (pendingLines.length) {
      const lines = pendingLines.slice();
      pendingLines = [];
      for (const l of lines) prettyPrint(l);
    }

    console.log(`Bienvenue ${myName} dans FLIP 7 !`);
    console.log("Tapez: READY pour commencer, QUIT pour quitter.");
  });

  socket.setEncoding("utf8");

  socket.on("data", (chunk) => {
    buffer += chunk;
    let idx;
    while ((idx = buffer.indexOf("\n")) !== -1) {
      const line = buffer.slice(0, idx).trim();
      buffer = buffer.slice(idx + 1);
      handleServerLine(line);
    }
  });

  socket.on("close", () => {
    console.log("Connexion fermée.");
    rl.close();
  });

  socket.on("error", (err) => {
    console.log("Erreur:", err.message);
    rl.close();
  });

  rl.on("line", (line) => {
    const msg = (line || "").trim();
    if (!msg) return;
    socket.write(msg + "\n");
  });
}

main().catch((e) => {
  console.log("Erreur client:", e.message);
  rl.close();
});
