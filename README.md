# coachproject — Ajout d'une ligne dans un Google Sheet et collecte LiveClient

Ce binaire scrute l'API LiveClient de League of Legends (https://127.0.0.1:2999/liveclientdata/allgamedata) en boucle, ne s'arrête jamais tant qu'il tourne, et vous assiste pour remplir un Google Sheet via:
- API Google Sheets (sécurisé, nécessite credentials), ou
- WebApp Apps Script (anonyme côté client, plus simple).

Il peut identifier votre champion grâce à votre RiotId fourni au lancement.

## Build Windows (.exe)

Sous Windows (cmd.exe):

```bat
cd C:\Users\Clement\GolandProjects\coachproject
go build -o coachproject.exe .
```

Alternative si vous avez Go 1.20+ et voulez préciser le répertoire:

```bat
go -C C:\Users\Clement\GolandProjects\coachproject build -o coachproject.exe .
```

Le fichier `coachproject.exe` sera généré à la racine du projet.

## Lancement — paramètres importants

- `--riotId "GameName#TagLine"` (obligatoire pour identifier votre champion parmi `allPlayers`).
  Exemple: `--riotId "XIN#ZHAO"`
- `--webapp "https://script.google.com/macros/s/XXXXX/exec"` (si vous utilisez le mode WebApp Apps Script)

Le programme:
- reste en boucle et n’exit jamais tout seul;
- ne vous demande pas si la partie est terminée tant que `IN_GAME` n’est pas `true` (i.e., tant qu’aucune donnée LiveClient n’a été reçue);
- quand le flux LiveClient s’arrête après avoir été `IN_GAME=true`, il vous pose des questions pour compléter les `values` (WIN/LOSS, ELO, MENTAL, TYPE_OF_GAME, commentaires, analyse, etc.).

### WebApp Apps Script

1) Créez un Apps Script et déployez-le en WebApp (voir `scripts/apps-script/Code.gs`).
   - Remplacez `SPREADSHEET_ID` et optionnellement `SHEET_GID`.
   - Deploy -> New deployment -> Web app -> Execute as: Me, Who has access: Anyone -> Deploy.
   - Copiez l’URL du WebApp.

2) Lancez l’appli avec votre RiotId et l’URL WebApp:

```bat
coachproject.exe --riotId "XIN#ZHAO" --webapp "https://script.google.com/macros/s/XXXXX/exec"
```

Le programme récupère votre champion en cherchant votre `riotId` dans `allPlayers`, calcule des métriques (CS/min, Deaths, KP), et vous pose les questions interactives à la fin de la partie:
- Lane jouée (TOP/JUNGLE/MIDDLE/BOTTOM/SUPPORT)
- Win ou Lose (tapez `w` ou `l`, la valeur envoyée sera `W` ou `L`)
- ELO avant la game (chaîne libre)
- Mental (tapez `1` pour `:D`, `2` pour `:|`, `3` pour `:c`)
- TYPE_OF_GAME (choisissez parmi `Free Win`, `Moyenne`, `Free Loose`)
- Post game commentary (texte libre)
- Analysis (texte libre)

3) Exemple d’envoi final des `values` (côté outil):
- `WIN_LOSS` (`W` ou `L`), `ELO`, `CHAMPION`, `LANE_GAME`, `DATE`, `MENTAL`, `POST_GAME_COMMENTARY`, `TYPE_OF_GAME`, `ANALYSIS`, `CS_M`, `DEATHS`, `KP`.