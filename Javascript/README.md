## Flip 7 — Jeu en mode texte

Implémentation du jeu Flip 7 en ligne de commande, jouable localement sur une seule machine.
Les joueurs jouent à tour de rôle via le terminal jusqu’à ce qu’un joueur atteigne 200 points.

## ▶️ Lancer le jeu

Depuis le dossier du projet :
```bash
node index.js
```


### Déroulement d’une partie

1. Le jeu demande le nombre de joueurs (minimum 2).
2. Une manche démarre automatiquement.
3. À chaque tour, un joueur :
- voit sa main,
- choisit de continuer ou de s’arrêter.
4. Les règles Flip 7 sont appliquées :
- doublon → élimination (sauf 2e chance),
- cartes spéciales,
- Flip 7 → fin immédiate de la manche.
5. Les scores sont calculés.
6. Une nouvelle manche commence.
7. Le premier joueur atteignant 200 points gagne.

### 💾 Sauvegarde

À la fin de chaque manche, les données sont enregistrées dans games.json
L’historique est réinitialisé au lancement d’une nouvelle partie.