package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"
)

const defaultWebAppURL = "https://script.google.com/macros/s/AKfycbwZrSwJAek8rQD3k0v2QZvBJNpurmZVEzVKxsdYQe8SaURmHRs7YklkHwomSxJXm8rj/exec"
const liveClientURL = "https://127.0.0.1:2999/liveclientdata/allgamedata"

// Structures minimales pour parser ce qu'on utilise
type LiveClientSnapshot struct {
	ActivePlayer struct {
		ChampionStats      struct{} `json:"championStats"`
		SummonerName       string   `json:"summonerName"`
		RiotID             string   `json:"riotId"`
		TeamRelativeColors bool     `json:"teamRelativeColors"`
	} `json:"activePlayer"`
	AllPlayers []struct {
		ChampionName string `json:"championName"`
		IsBot        bool   `json:"isBot"`
		Position     string `json:"position"`
		Team         string `json:"team"`
		SummonerName string `json:"summonerName"`
		RiotID       string `json:"riotId"`
		Scores       struct {
			Assists    int     `json:"assists"`
			CreepScore int     `json:"creepScore"`
			Deaths     int     `json:"deaths"`
			Kills      int     `json:"kills"`
			WardScore  float64 `json:"wardScore"`
		} `json:"scores"`
	} `json:"allPlayers"`
	GameData struct {
		GameMode   string  `json:"gameMode"`
		GameTime   float64 `json:"gameTime"` // en secondes
		MapName    string  `json:"mapName"`
		MapNumber  int     `json:"mapNumber"`
		MapTerrain string  `json:"mapTerrain"`
	} `json:"gameData"`
}

func main() {
	timeout := flag.Duration("timeout", 15*time.Second, "Timeout HTTP")
	pollEvery := flag.Duration("poll", 2*time.Second, "Interval de polling Live Client")
	riotIDArg := flag.String("riotid", "", "Votre Riot ID")
	webappArg := flag.String("webapp", defaultWebAppURL, "URL du WebApp à utiliser")
	flag.Parse()

	webappURL := *webappArg

	reader := bufio.NewReader(os.Stdin)

	// État de la partie
	inGame := false
	lastStatus := "" // "waiting", "ingame", "disconnected"

	// Boucle infinie: ne quitte jamais, relance le polling après chaque interaction
	for {
		// Polling Live Client jusqu'à interruption
		snapshot, hadData, err := pollLiveClient(*pollEvery, *timeout)

		// Gestion des erreurs pour éviter le spam
		if err != nil {
			if hadData {
				// Déconnexion après avoir eu des données -> on traite normalement
				_, _ = fmt.Fprintf(os.Stderr, "Arrêt du polling: %v\n", err)
			} else {
				// Pas encore de partie: ne pas spammer l'erreur, juste afficher l'état une fois et attendre
				if lastStatus != "waiting" {
					fmt.Println("Toujours en attente de début de partie…")
					lastStatus = "waiting"
				}
				// Petit délai avant de relancer le polling
				time.Sleep(2 * time.Second)
				continue
			}
		}

		// Si on a reçu des données au moins une fois, on est en game
		if hadData {
			inGame = true
			if lastStatus != "ingame" {
				fmt.Println("Partie détectée: passage en mode IN_GAME.")
				lastStatus = "ingame"
			}
		} else {
			// Jamais reçu de données -> ne pas demander fin de partie (message déjà géré ci-dessus)
			if lastStatus != "waiting" {
				fmt.Println("Toujours en attente de début de partie…")
				lastStatus = "waiting"
			}
		}

		// On ne pose la question de fin de partie que si on était en game
		if inGame {
			if lastStatus != "disconnected" {
				fmt.Println("Le Live Client ne répond plus.")
				lastStatus = "disconnected"
			}
			fmt.Print("Est-ce que la partie est terminée ? (oui/non): ")
			resp, _ := reader.ReadString('\n')
			resp = strings.ToLower(strings.TrimSpace(resp))

			if resp == "oui" || resp == "o" || resp == "yes" || resp == "y" {
				// La partie est terminée -> repasser inGame à false après l'envoi
				data := buildDataFromSnapshot(snapshot, *riotIDArg)

				// Demander la lane jouée (avant Win/Lose)
				fmt.Println("Lane jouée : TOP (t), JUNGLE (j), MIDDLE (m), BOTTOM (b), SUPPORT (s)")
				fmt.Print("Choisissez votre lane (t/j/m/b/s ou nom complet): ")
				laneInput, _ := reader.ReadString('\n')
				laneInput = strings.ToLower(strings.TrimSpace(laneInput))
				switch laneInput {
				case "t", "top":
					data.Set("LANE_GAME", "TOP")
				case "j", "jungle":
					data.Set("LANE_GAME", "JUNGLE")
				case "m", "mid", "middle":
					data.Set("LANE_GAME", "MIDDLE")
				case "b", "bot", "bottom":
					data.Set("LANE_GAME", "BOTTOM")
				case "s", "sup", "support":
					data.Set("LANE_GAME", "SUPPORT")
				default:
					fmt.Println("Entrée invalide pour la lane, valeur détectée auto conservée.")
				}

				// Poser les questions pour compléter les values
				fmt.Print("Résultat de la game (w = Win, l = Lose): ")
				wl, _ := reader.ReadString('\n')
				wl = strings.ToLower(strings.TrimSpace(wl))
				if wl == "w" {
					data.Set("WIN_LOSS", "W")
				} else if wl == "l" {
					data.Set("WIN_LOSS", "L")
				} else {
					fmt.Println("Entrée invalide pour Win/Lose, valeur laissée vide.")
				}

				fmt.Print("Votre ELO avant la game (ex: D4, P1, Gold 2): ")
				elo, _ := reader.ReadString('\n')
				elo = strings.TrimSpace(elo)
				if elo != "" {
					data.Set("ELO", elo)
				}

				// Question MENTAL
				fmt.Println("Votre mental pendant la game : 1 = :D, 2 = :|, 3 = :c")
				fmt.Print("Choisissez (1/2/3): ")
				mentalChoice, _ := reader.ReadString('\n')
				mentalChoice = strings.TrimSpace(mentalChoice)
				switch mentalChoice {
				case "1":
					data.Set("MENTAL", ":D")
				case "2":
					data.Set("MENTAL", ":|")
				case "3":
					data.Set("MENTAL", ":c")
				default:
					fmt.Println("Entrée invalide pour MENTAL, valeur laissée vide.")
				}

				// Post-game commentary (texte libre)
				fmt.Print("Commentaire post-game (votre ressenti): ")
				pgc, _ := reader.ReadString('\n')
				pgc = strings.TrimSpace(pgc)
				if pgc != "" {
					data.Set("POST_GAME_COMMENTARY", pgc)
				}

				// Analysis (texte libre)
				fmt.Print("Analyse (points clés, à améliorer): ")
				analysis, _ := reader.ReadString('\n')
				analysis = strings.TrimSpace(analysis)
				if analysis != "" {
					data.Set("ANALYSIS", analysis)
				}

				// Type of game (3 options)
				fmt.Println("Type de game : 1 = Free Win, 2 = Moyenne, 3 = Free Loose")
				fmt.Print("Choisissez (1/2/3): ")
				togChoice, _ := reader.ReadString('\n')
				togChoice = strings.TrimSpace(togChoice)
				switch togChoice {
				case "1":
					data.Set("TYPE_OF_GAME", "Free Win")
				case "2":
					data.Set("TYPE_OF_GAME", "Moyenne")
				case "3":
					data.Set("TYPE_OF_GAME", "Free Loose")
				default:
					fmt.Println("Entrée invalide pour TYPE_OF_GAME, valeur laissée vide.")
				}

				// Affiche les données finales avant envoi
				fmt.Println("Données finales pour envoi:", data.Encode())

				if err := postToWebApp(webappURL, data, *timeout); err != nil {
					_, _ = fmt.Fprintf(os.Stderr, "Échec WebApp: %v\n", err)
				} else {
					fmt.Println("Succès: données envoyées via WebApp.")
				}

				// Partie terminée -> reset IN_GAME et statut
				inGame = false
				lastStatus = "waiting"
			} else {
				// Il n'a pas fini -> on reste IN_GAME
				fmt.Println("OK, la partie n'est pas terminée, on continue la surveillance…")
			}
		}

		// Relance la boucle: on repart en polling sans quitter
	}
}

// Construit le url.Values en utilisant les infos disponibles dans le snapshot
func buildDataFromSnapshot(s LiveClientSnapshot, riotID string) url.Values {
	champName := ""
	lane := ""
	var selfScores struct{ Kills, Assists, Deaths, CS int }

	// Déterminer le joueur cible: par riotID si fourni, sinon via SummonerName
	matchBy := func(pSummoner, pRiot string) bool {
		if riotID != "" {
			return strings.EqualFold(pRiot, riotID)
		}
		return strings.EqualFold(pSummoner, s.ActivePlayer.SummonerName)
	}

	for _, p := range s.AllPlayers {
		if matchBy(p.SummonerName, p.RiotID) {
			champName = p.ChampionName
			lane = p.Position
			selfScores.Kills = p.Scores.Kills
			selfScores.Assists = p.Scores.Assists
			selfScores.Deaths = p.Scores.Deaths
			selfScores.CS = p.Scores.CreepScore
			break
		}
	}
	if champName == "" {
		// fallback: prendre premier non-bot
		for _, p := range s.AllPlayers {
			if !p.IsBot {
				champName = p.ChampionName
				lane = p.Position
				selfScores.Kills = p.Scores.Kills
				selfScores.Assists = p.Scores.Assists
				selfScores.Deaths = p.Scores.Deaths
				selfScores.CS = p.Scores.CreepScore
				break
			}
		}
	}

	// Normaliser la lane: l'API retourne parfois "UTILITY" pour support
	if strings.EqualFold(lane, "UTILITY") {
		lane = "SUPPORT"
	}

	// CS par minute
	minutes := s.GameData.GameTime / 60.0
	csPerMin := 0.0
	if minutes > 0 {
		csPerMin = float64(selfScores.CS) / minutes
	}
	// Arrondir à une décimale
	csPerMinStr := fmt.Sprintf("%.1f", csPerMin)

	// Deaths
	deathsStr := fmt.Sprintf("%d", selfScores.Deaths)

	// Kill participation et rang (1 = meilleure KP de l'équipe)
	// Utiliser riotID pour identifier selfName si disponible
	selfName := s.ActivePlayer.SummonerName
	if riotID != "" {
		selfName = riotID
	}
	kpRank := computeKPRankByID(s, selfName, riotID != "")
	kpRankStr := fmt.Sprintf("%d", kpRank)

	// Date au format YYYY-MM-DD
	dateStr := time.Now().Format("2006-01-02")

	values := url.Values{
		"WIN_LOSS":             {""}, // non déterminé ici
		"ELO":                  {""}, // non disponible
		"CHAMPION":             {champName},
		"LANE_GAME":            {lane},
		"DATE":                 {dateStr},
		"MENTAL":               {""}, // non disponible
		"POST_GAME_COMMENTARY": {""}, // non disponible
		"TYPE_OF_GAME":         {""}, // non disponible
		"ANALYSIS":             {""}, // non disponible
		"CS_M":                 {csPerMinStr},
		"DEATHS":               {deathsStr},
		"KP":                   {kpRankStr},
	}

	// Afficher en console ce qui sera envoyé (encodé en x-www-form-urlencoded)
	fmt.Println("Données construites pour envoi:", values.Encode())

	return values
}

// Calcule le rang de kill participation (1 = meilleur) pour l'équipe
// Si useRiotID est vrai, selfName représente un riotId et la comparaison se fait sur RiotID
func computeKPRankByID(s LiveClientSnapshot, selfName string, useRiotID bool) int {
	// Déterminer l'équipe du joueur actif
	team := ""
	for _, p := range s.AllPlayers {
		if useRiotID {
			if strings.EqualFold(p.RiotID, selfName) {
				team = p.Team
				break
			}
		} else {
			if strings.EqualFold(p.SummonerName, selfName) {
				team = p.Team
				break
			}
		}
	}
	if team == "" {
		return 0
	}
	// Somme des kills de l'équipe
	teamKills := 0
	for _, p := range s.AllPlayers {
		if p.Team == team {
			teamKills += p.Scores.Kills
		}
	}
	if teamKills == 0 {
		return 0
	}
	// KP pour chaque joueur de l'équipe
	type pair struct {
		name string
		kp   float64
	}
	var kps []pair
	for _, p := range s.AllPlayers {
		if p.Team == team {
			kp := float64(p.Scores.Kills+p.Scores.Assists) / float64(teamKills)
			id := p.SummonerName
			if useRiotID {
				id = p.RiotID
			}
			kps = append(kps, pair{name: id, kp: kp})
		}
	}
	// Trier desc
	sort.Slice(kps, func(i, j int) bool { return kps[i].kp > kps[j].kp })
	// Trouver le rang
	for i, pr := range kps {
		if strings.EqualFold(pr.name, selfName) {
			return i + 1
		}
	}
	return 0
}

// Modifie la signature pour renvoyer si on a déjà reçu des données
func pollLiveClient(interval, timeout time.Duration) (LiveClientSnapshot, bool, error) {
	// Client HTTPS vers 127.0.0.1 avec cert non vérifié (Live Client utilise un cert local)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // OK pour localhost Live Client
	}
	client := &http.Client{
		Timeout:   timeout,
		Transport: tr,
	}

	var last LiveClientSnapshot
	hadData := false

	for {
		req, err := http.NewRequest(http.MethodGet, liveClientURL, nil)
		if err != nil {
			return last, hadData, err
		}

		resp, err := client.Do(req)
		if err != nil {
			// Stoppe la boucle quand ça ne répond plus
			return last, hadData, fmt.Errorf("Live Client non joignable: %w", err)
		}

		// Lecture et affichage du body
		func() {
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				_, _ = fmt.Fprintf(os.Stderr, "Statut non OK (%d), arrêt du polling.\n", resp.StatusCode)
				return
			}
			b, readErr := io.ReadAll(resp.Body)
			if readErr != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Erreur de lecture du body: %v\n", readErr)
				return
			}
			// Affiche le body renvoyé par l'API Live Client
			fmt.Println("LiveClient body:", string(b))
			// Parser et conserver le dernier snapshot
			var tmp LiveClientSnapshot
			if err := json.Unmarshal(b, &tmp); err == nil {
				last = tmp
				hadData = true
			}
		}()

		// Continue à poll avec une pause
		time.Sleep(interval)
	}
}

func postToWebApp(webappURL string, data url.Values, timeout time.Duration) error {
	reqBody := strings.NewReader(data.Encode())

	req, err := http.NewRequest(http.MethodPost, webappURL, reqBody)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("statut HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	return nil
}

func failf(format string, a ...any) {
	// Ne quitte jamais le programme; juste log
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", a...)
}
