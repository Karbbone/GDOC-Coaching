// Apps Script WebApp pour ajouter une nouvelle ligne avec la valeur en colonne A
// Déployez en tant que Web App: Exécuter en tant que: Moi (ou le compte ayant accès), Accès: Anyone

// Renseignez ici l'ID du spreadsheet et (optionnel) le GID de l'onglet cible
const SPREADSHEET_ID = '1q7xSp5B6eFEvrPUzM4_K-FfNcgPMSwY5fp3iPJSQ6KI';
const SHEET_GID = 1845885794; // mettez null si vous voulez le premier onglet

function doPost(e) {
  try {
    const value = (e && e.parameter && e.parameter.value) ? e.parameter.value : 'W';
    const ss = SpreadsheetApp.openById(SPREADSHEET_ID);
    const sheet = resolveSheet(ss, SHEET_GID);
    sheet.appendRow([value]);
    return ContentService.createTextOutput('OK');
  } catch (err) {
    return ContentService.createTextOutput('ERR: ' + err).setMimeType(ContentService.MimeType.TEXT);
  }
}

function doGet(e) {
  // Support GET pour tests rapides ?value=...
  return doPost(e);
}

function resolveSheet(ss, gid) {
  if (!gid && gid !== 0) {
    return ss.getSheets().sort((a, b) => a.getIndex() - b.getIndex())[0];
  }
  const sheets = ss.getSheets();
  for (var i = 0; i < sheets.length; i++) {
    if (sheets[i].getSheetId() === gid) return sheets[i];
  }
  // Fallback: premier onglet si pas trouvé
  return sheets.sort((a, b) => a.getIndex() - b.getIndex())[0];
}

