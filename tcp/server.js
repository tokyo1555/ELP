// server.js
const net = require("net");
const os = require("os");

const PORT = 5000;
const HOST = "0.0.0.0";

function getWifiIPv4() {
  const ifaces = os.networkInterfaces();
  const wifiRegex = /(wi-?fi|wlan|wireless|réseau sans fil)/i;

  // Priorité Wi-Fi
  for (const name of Object.keys(ifaces)) {
    if (!wifiRegex.test(name)) continue;
    for (const it of ifaces[name]) {
      if (it.family === "IPv4" && !it.internal) return { ip: it.address, iface: name };
    }
  }

  // Fallback
  for (const name of Object.keys(ifaces)) {
    for (const it of ifaces[name]) {
      if (it.family === "IPv4" && !it.internal) return { ip: it.address, iface: name };
    }
  }

  return { ip: "IP introuvable", iface: "unknown" };
}

function sendLine(sock, line) {
  sock.write(line + "\n");
}

const players = []; // { id, socket, name, ready }

function broadcast(line) {
  players.forEach((p) => sendLine(p.socket, line));
}

function lobbyStatus() {
  const readyCount = players.filter((p) => p.ready).length;
  broadcast(`LOBBY ${players.length} joueurs | READY ${readyCount}/${players.length}`);
  broadcast(
    "JOUEURS " +
      players
        .map((p) => `${p.name || "J" + p.id}:${p.ready ? "READY" : "WAIT"}`)
        .join(" | ")
  );
}

// ID compact (réutilise les IDs libérés)
function getFreeId() {
  const used = new Set(players.map((p) => p.id));
  let id = 1;
  while (used.has(id)) id++;
  return id;
}

function tryStart() {
  if (players.length >= 2 && players.every((p) => p.ready)) {
    broadcast("GAME_START");
    broadcast(`NB_JOUEURS ${players.length}`);
    broadcast("INFO Début de partie (moteur Flip7 à brancher ici)");
  }
}

function labelPlayer(p) {
  return `Joueur num ${p.id}`;
}

const server = net.createServer((socket) => {
  socket.setEncoding("utf8");
  socket.setNoDelay(true);

  const id = getFreeId();
  const player = { id, socket, name: null, ready: false };
  players.push(player);

  // Affichage serveur demandé (événement: connexion)
  console.log(`${labelPlayer(player)} connecté`);

  // Protocole côté client
  sendLine(socket, `WELCOME J${id}`);
  sendLine(socket, "CMD: NAME <pseudo>");
  sendLine(socket, "Puis: READY | QUIT");

  // On peut envoyer l'état lobby aux clients
  lobbyStatus();

  let buffer = "";

  socket.on("data", (chunk) => {
    buffer += chunk;
    let idx;

    while ((idx = buffer.indexOf("\n")) !== -1) {
      const line = buffer.slice(0, idx).trim();
      buffer = buffer.slice(idx + 1);
      if (!line) continue;

      const upper = line.toUpperCase();

      // NAME <pseudo>
      if (upper.startsWith("NAME ")) {
        const name = line.slice(5).trim();
        if (!name) {
          sendLine(socket, "ERR Nom invalide. Utilise: NAME <pseudo>");
          continue;
        }

        player.name = name;

        // Affichage serveur demandé (événement: nom défini)
        console.log(
          `${labelPlayer(player)} s'appelle ${player.name} (${player.ready ? "Ready" : "Wait"})`
        );

        broadcast(`INFO J${player.id} s'appelle ${name}`);
        lobbyStatus();
        continue;
      }

      // READY
      if (upper === "READY") {
        if (!player.name) {
          sendLine(socket, "ERR Choisis d'abord un nom avec: NAME <pseudo>");
          continue;
        }

        player.ready = true;

        // Affichage serveur demandé (événement: ready)
        console.log(`${labelPlayer(player)} s'appelle ${player.name} (Ready)`);

        broadcast(`INFO ${player.name} est READY`);
        lobbyStatus();
        tryStart();
        continue;
      }

      // QUIT
      if (upper === "QUIT") {
        sendLine(socket, "BYE");
        socket.end();
        continue;
      }

      sendLine(socket, "ERR Commande inconnue. Utilise: NAME <pseudo> | READY | QUIT");
    }
  });

  function removePlayer() {
    const i = players.indexOf(player);
    if (i !== -1) players.splice(i, 1);

    const who = player.name ? `${labelPlayer(player)} (${player.name})` : labelPlayer(player);
    console.log(`${who} déconnecté`);

    broadcast(`INFO ${player.name || "J" + player.id} a quitté`);
    lobbyStatus();
  }

  socket.on("close", () => {
    removePlayer();
  });

  socket.on("error", (err) => {
    // Si fermeture brutale côté client (ECONNRESET), close arrivera aussi.
    // On log seulement l'erreur sans spam.
    console.log(`Erreur socket ${player.name || "J" + player.id} : ${err.message}`);
  });
});

server.listen(PORT, HOST, () => {
  const { ip, iface } = getWifiIPv4();
  console.log("Serveur TCP lancé");
  console.log(`Interface utilisée : ${iface}`);
  console.log(`Adresse à donner aux joueurs : ${ip}:${PORT}`);
  console.log("En attente de joueurs...");
});

server.on("error", (err) => {
  console.log("Erreur serveur :", err.message);
});
