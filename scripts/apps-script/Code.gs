// Apps Script WebApp pour ajouter une nouvelle ligne avec toutes les valeurs envoyées
// Déployez en tant que Web App: Exécuter en tant que: Moi, Accès: Anyone

const SPREADSHEET_ID = 'A REMPLIR';
const SHEET_GID = A REMPLIR; // sinon null pour premier onglet

function doPost(e) {
  try {
    const p = e.parameter;

    // Lecture des champs envoyés depuis Go
    const row = [
      p.WIN_LOSS || "",
      p.ELO || "",
      p.CHAMPION || "",
      "",
      p.LANE_GAME || "",
      p.DATE || "",
      p.MENTAL || "",
      p.POST_GAME_COMMENTARY || "",
      p.TYPE_OF_GAME || "",
      p.ANALYSIS || "",
      p.CS_M || "",
      p.DEATHS || "",
      p.KP || ""
    ];

    const ss = SpreadsheetApp.openById(SPREADSHEET_ID);
    const sheet = resolveSheet(ss, SHEET_GID);
    sheet.appendRow(row);

    return ContentService.createTextOutput('OK');
  } catch (err) {
    return ContentService
      .createTextOutput('ERR: ' + err)
      .setMimeType(ContentService.MimeType.TEXT);
  }
}

function doGet(e) {
  return doPost(e);
}

function resolveSheet(ss, gid) {
  if (!gid && gid !== 0) {
    return ss.getSheets()[0];
  }
  const sheets = ss.getSheets();
  for (var i = 0; i < sheets.length; i++) {
    if (sheets[i].getSheetId() === gid) {
      return sheets[i];
    }
  }
  return ss.getSheets()[0];
}
