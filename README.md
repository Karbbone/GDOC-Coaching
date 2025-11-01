# coachproject — Ajout d'une ligne "W" dans un Google Sheet

Deux modes au choix:

1) API Google Sheets (sécurisé, nécessite credentials)
- Créez un projet GCP + un Compte de service (JSON) avec l'API Google Sheets activée.
- Partagez la feuille avec l'email du compte de service (rôle: Éditeur).
- Exécutez l'appli:

```bat
set GOOGLE_APPLICATION_CREDENTIALS=C:\chemin\vers\service-account.json
coachproject.exe --url "https://docs.google.com/spreadsheets/d/1q7xSp5B6eFEvrPUzM4_K-FfNcgPMSwY5fp3iPJSQ6KI/edit?gid=1845885794#gid=1845885794" --value W
```

ou

```bat
coachproject.exe --credentials C:\chemin\vers\service-account.json --url "https://docs.google.com/spreadsheets/d/1q7xSp5B6eFEvrPUzM4_K-FfNcgPMSwY5fp3iPJSQ6KI/edit?gid=1845885794#gid=1845885794" --value W
```

Options utiles:
- `--sheetTitle "Nom d'onglet"` pour cibler un onglet spécifique.
- `--timeout 30s` pour ajuster le timeout.

2) WebApp Apps Script (anonyme, pas de credentials côté client)
- Ouvrez Google Drive -> Nouveau -> Apps Script.
- Collez le contenu de `scripts/apps-script/Code.gs`.
- Remplacez `SPREADSHEET_ID` et (optionnel) `SHEET_GID` par vos valeurs.
- Déployez: Deploy -> New deployment -> Web app -> Execute as: Me, Who has access: Anyone -> Deploy.
- Récupérez l'URL du WebApp.
- Exécutez l'appli:

```bat
coachproject.exe --webapp "https://script.google.com/macros/s/XXXXX/exec" --value W
```

Build local (Windows, cmd.exe):

```bat
go -C C:\Users\Clement\GolandProjects\coachproject build .
```

Notes:
- Même si la feuille est ouverte en édition via lien, l'API requiert une identité. D'où les deux modes.
- Par défaut, la valeur écrite est `W`. Vous pouvez changer via `--value`.

