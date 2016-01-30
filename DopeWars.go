package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"encoding/json"
	"math/rand"
	"time"
	"strconv"

	"github.com/Pallinder/go-randomdata"
	"github.com/boltdb/bolt"
	"log"
	"github.com/dustin/go-humanize"
)

const NUMBER_OF_LOCATIONS = 5
const NUMBER_OF_ENCOUNTER = 2
//var templates = template.Must(template.ParseGlob("templates/*" ))

type ItemPriceRanges struct {
	Min int
	Max int
}
var priceRanges map[string]ItemPriceRanges

type Location struct {
	Name  		string
	Attributes	string
}
var Locs [NUMBER_OF_LOCATIONS]Location

type Encounter struct {
	Name  		string
	Text 		string
	Chance		int
	Damage		int
	Cost		int
}
var Encounters [NUMBER_OF_ENCOUNTER]Encounter

type Page struct {
	Title 		string
	Game 		Game
	Locations 	[NUMBER_OF_LOCATIONS]Location
}

type Game struct {
	Name      	string
	Health 		int
	Year		int
	Location 	string
	PercentDone	int
	Cash		int
	Bank		int
	Debt 		int
	Worth 		int
	Messages	[]string
	Prices 		map[string]int
	Inventory 	map[string]int
	Ended		bool
}

func (g *Game) InsertMessage (msg string) {
	g.Messages = append([]string{msg}, g.Messages...)
}

var db *bolt.DB

func createSessionId() string {
	sid := randomdata.StringNumberExt(4, "-", 4)
	newGame := Game{
		Name: randomdata.SillyName(),
		Health: 100,
		Year: 1,
		Location: Locs[0].Name,
		Cash: 2000,
		Debt: 5500,
		Bank: 0,
		Worth: -9000,
		Ended: false,
		Messages: make([]string, 1),
	}
	newGame.InsertMessage("New Game Started!!!\r")
	newGame.Prices = initialDrugPrices()
	newGame.Inventory = make(map[string]int)
	fmt.Println(newGame)
	storeGame(sid, newGame)
	return sid
}

func getUserId(w http.ResponseWriter, r *http.Request) string {
	var sid string
	cookie, err := r.Cookie("DopeSessionId")
	if err != nil {
		sid = createSessionId()
		cookie := http.Cookie{Name: "DopeSessionId", Value: sid, Path: "/", MaxAge: (60*60*24*365)}
		http.SetCookie(w, &cookie)
	} else {
		sid, _ = url.QueryUnescape(cookie.Value)
	}
	return sid
}


func renderTemplate(w http.ResponseWriter, tmpl string, p Page) {
	if p.Game.Ended {
		renderEndgame(w, "end_game", p)
		return
	}
	templateFuncs := template.FuncMap{
		"humanizeComma": func(s int) string {
			i := int64(s)
			return "$ " + humanize.Comma(i) + ".00"
		},
	}
	p.Locations = Locs
	templates := template.New("templates").Funcs(templateFuncs)
	templates = template.Must(templates.ParseFiles("index.html"))
	templates = template.Must(templates.ParseGlob("templates/*"))
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func renderEndgame(w http.ResponseWriter, tmpl string, p Page) {
	templateFuncs := template.FuncMap{
		"humanizeComma": func(s int) string {
			i := int64(s)
			return "$ " + humanize.Comma(i) + ".00"
		},
	}
	p.Locations = Locs
	templates := template.New("templates").Funcs(templateFuncs)
	templates = template.Must(templates.ParseFiles("end_game.html"))
	templates = template.Must(templates.ParseGlob("templates/*"))
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("in default handler")
	sid := getUserId(w, r)
	gam := retrieveGame(sid)
	fmt.Println(gam)
	if sid == "" || gam.Name == "" {
		sid = createSessionId()
		cookie := http.Cookie{Name: "DopeSessionId", Value: url.QueryEscape(sid), Path: "/", HttpOnly: true, MaxAge: int(3600)}
		http.SetCookie(w, &cookie)
		gam = retrieveGame(sid)
	}
	g := Page{Title: "Home", Game: gam }
	renderTemplate(w, "index", g)
}

func newgameHandler(w http.ResponseWriter, r *http.Request) {
	sid := getUserId(w, r)
	deleteGame(sid)
	sid = createSessionId()
	cookie := http.Cookie{Name: "DopeSessionId", Value: url.QueryEscape(sid), Path: "/", HttpOnly: true, MaxAge: int(3600)}
	http.SetCookie(w, &cookie)
	g := retrieveGame(sid)
	p := Page{Title: "Home", Game: g}
	renderTemplate(w, "index", p)
}

func ProcessTurn(loc string, sid string, gam Game) Game {
	log.Println(gam.Location, loc, sid)
	if gam.Location == loc {
		return gam
	}

	// We are moving
	gam.Year++
	gam.Worth = gam.Cash + gam.Bank - (gam.Debt * 2)
	log.Println("Game Worth:", gam.Worth, "Cash:", gam.Cash, "Bank:", gam.Bank, "Debt:", gam.Debt)
	if gam.Year > 30 {
		gam.Year--
		gam.Ended = true
		storeGame(sid, gam)
		return gam
	}
	gam.Prices = updateDrugPrices(gam.Prices)
	if gam.Health < 100 {
		gam.Health++
	}
	gam.Location = loc
	gam.PercentDone = int(float64(gam.Year) / float64(30) * float64(100))
	gam.InsertMessage(fmt.Sprintf("Took a Cab to %s\r", loc))

	done := false
	for _, encounter := range Encounters {
		// Encounter
		log.Println(encounter.Name)
		cost := 0
		damage := 0
		if encounter.Cost > 0 {
			cost = random(0, encounter.Cost)
		}
		if encounter.Damage > 0 {
			damage = random(0, encounter.Damage)
		}
		switch(encounter.Name) {
		case "Mugging":
			if encounter.Chance >= random(0, 100) {
				gam.Cash = gam.Cash - cost
				gam.Health = gam.Health - damage
				str := fmt.Sprintf(encounter.Text, cost, damage);
				gam.InsertMessage(fmt.Sprintf("As you arrive at %s you %s\n", loc, str))
				done = true
			}
		case "Police":
			if encounter.Chance >= random(0, 100) {
				gam.Cash = gam.Cash - cost
				gam.Health = gam.Health - damage
				str := fmt.Sprintf(encounter.Text, random(0, encounter.Damage));
				gam.InsertMessage(fmt.Sprintf("As you arrive at %s you %s\n", loc, str))
				done = true
			}
		}
		if done == true {
			break
		}
	}

	storeGame(sid, gam)
	return gam
}

func cabHandler(w http.ResponseWriter, r *http.Request) {
	loc := r.FormValue("loc")
	sid := getUserId(w, r)
	var gam = retrieveGame(sid)
	gam = ProcessTurn(loc, sid, gam)

	g := Page{Title: "Home", Game: gam}
	renderTemplate(w, "index", g)
}

func buyHandler(w http.ResponseWriter, r *http.Request) {
	sid := getUserId(w, r)
	var gam = retrieveGame(sid)
	drug := r.FormValue("drug")
	quant, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil {
		log.Println("Quantity from From could not be converted to an int:", r.FormValue("quantity"))
	} else {
		price := gam.Prices[drug] * quant
		log.Println("Drug Price:", price)
		if(price <= gam.Cash) {
			gam.Inventory[drug] = gam.Inventory[drug] + quant
			gam.Cash = gam.Cash - price
			log.Println("Success", gam.Name, "purchased", quant, "of", drug)
			gam.InsertMessage(fmt.Sprintf("Purchased %d of %s ,price: %d\r", quant, drug, price))
			gam.Worth = gam.Cash + gam.Bank - (gam.Debt * 2)
			log.Println("Game Worth:", gam.Worth, "Cash:", gam.Cash, "Bank:", gam.Bank, "Debt:", gam.Debt)
			storeGame(sid, gam)
		}
	}
	g := Page{Title: "Home", Game: gam}
	renderTemplate(w, "index", g)
}

func sellHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("In Selling Handler")
	sid := getUserId(w, r)
	var gam = retrieveGame(sid)
	drug := r.FormValue("drug")
	quant, err := strconv.Atoi(r.FormValue("amount"))
	if err != nil {
		log.Println("Quantity from From could not be converted to an int:", r.FormValue("quantity"))
	} else {
		price := gam.Prices[drug] * quant
		log.Println("Drug Quantity to sell:", quant)
		if(quant <= gam.Inventory[drug]) {
			gam.Inventory[drug] = gam.Inventory[drug] - quant
			if gam.Inventory[drug] == 0 {
				delete(gam.Inventory, drug)
			}
			gam.Cash = gam.Cash + price
			log.Println("Success", gam.Name, "purchased", quant, "of", drug)
			gam.InsertMessage(fmt.Sprintf("Sold %d of %s ,price: %d\r", quant, drug, price))
			gam.Worth = gam.Cash + gam.Bank - (gam.Debt * 2)
			log.Println("Game Worth:", gam.Worth, "Cash:", gam.Cash, "Bank:", gam.Bank, "Debt:", gam.Debt)
			storeGame(sid, gam)
		}
	}
	g := Page{Title: "Home", Game: gam}
	renderTemplate(w, "index", g)
}

func random(min, max int) int {
	return rand.Intn(max - min) + min
}

func updateDrugPrices(gamePrices map[string]int) (map[string]int) {
	p := make(map[string]int)

//	for _, loc := range locationNames {
	for drug, price := range priceRanges {
		TenPercent := ((price.Max - price.Min) /10 * 100)
		min := (gamePrices[drug] - TenPercent)
		if min < priceRanges[drug].Min  {
			min = priceRanges[drug].Min
		}
		max := (gamePrices[drug] + TenPercent)
		if max > priceRanges[drug].Max {
			max = priceRanges[drug].Max
		}
		p[drug] = random(min, max)
	}
//	}
	log.Println(p)
	return p
}

func initialDrugPrices() map[string]int {
	p := make(map[string]int)

//	for _, loc := range locationNames {
		for drug, price := range priceRanges {
			p[drug] = random(price.Min, price.Max)
		}
//	}
	log.Println(p)
	return p
}

func init() {
	rand.Seed(time.Now().Unix())
	priceRanges = make(map[string]ItemPriceRanges)
	priceRanges["Acid"] = ItemPriceRanges{Min:1000, Max: 4400}
	priceRanges["Cocaine"] = ItemPriceRanges{Min:15000, Max: 29000}
	priceRanges["Hashish"] = ItemPriceRanges{Min:480, Max: 1280}
	priceRanges["Heroin"] = ItemPriceRanges{Min:5500, Max: 13000}
	priceRanges["Ludes"] = ItemPriceRanges{Min:11, Max: 60}
	priceRanges["Meth"] = ItemPriceRanges{Min:1500, Max: 4400}
	priceRanges["Peyote"] = ItemPriceRanges{Min:220, Max: 700}
	priceRanges["Shrooms"] = ItemPriceRanges{Min:630, Max: 1300}
	priceRanges["Speed"] = ItemPriceRanges{Min:90, Max: 250}
	priceRanges["Weed"] = ItemPriceRanges{Min:315, Max: 890}
	priceRanges["Opium"] = ItemPriceRanges{Min:540, Max: 1250}
	priceRanges["PCP"] = ItemPriceRanges{Min:1000, Max: 2500}

	Locs[0] = Location{Name: "The Bronks", Attributes: "Loan Shark"}
	Locs[1] = Location{Name: "The Ghetto", Attributes: "Guns"}
	Locs[2] = Location{Name: "Central Park", Attributes: "Hospital"}
	Locs[3] = Location{Name: "Manhattan", Attributes: "Bank"}
	Locs[4] = Location{Name: "Coney Island", Attributes: "-"}

	Encounters[0] = Encounter{Name: "Mugging", Text: "are mugged by a band of thieves and they steal %d dollars, and you lose %d health.", Chance: 10, Damage: 10, Cost: 50}
	Encounters[1] = Encounter{Name: "Police", Text: "are  greeted by the cops.  After a struggle you escaped but lost %d health.", Chance: 10, Damage: 10, Cost: 0}
}

func main() {
	var err error
	db, err = bolt.Open("games.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fs11 := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs11))
	http.HandleFunc("/", defaultHandler)
	http.HandleFunc("/cab", cabHandler)
	http.HandleFunc("/buy", buyHandler)
	http.HandleFunc("/sell", sellHandler)

	http.HandleFunc("/delete", newgameHandler)
	http.ListenAndServe(":8080", nil)
}

func storeGame(sid string, gam Game) {
	g, err := json.Marshal(gam)
	if err != nil {
		log.Fatalln("Could not marshall", err)
		return
	}

	set(sid, string(g))
}

func retrieveGame(sid string) Game {
	var bg Game
	val, err := get(sid)
	if err != nil || val == "" {
		log.Println("Could not retrieve", err)
		return Game{}
	}
	bg = Game{}
	err = json.Unmarshal([]byte(val), &bg)
	if err != nil {
		log.Fatalln("Error Unmarshal", err)
	}
	return bg
}

func deleteGame(sid string) {
	db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("games"))
		if err != nil {
			return err
		}
		b.Delete([]byte(sid))
		return nil
	})
}


func set(key string, val string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("games"))
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(key), []byte(val))
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}


func get(key string) (string, error) {
	var val []byte

	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("games"))
		if bucket == nil {
			return fmt.Errorf("Bucket %q not found!", []byte("games"))
		}

		val = bucket.Get([]byte(key))

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
	return string(val), err
}
