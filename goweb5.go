package main
import (
  "fmt"
  "net/http"
  "html/template"
  "io/ioutil"
  "encoding/xml"
  "sync"
)

var wg sync.WaitGroup

type Newsmap struct{
  PublicationDate string
  Location string
}

type NewsAggPage struct{
  Title string
  News map[string] Newsmap
}


type Sitemapindex struct{
  Locations []string `xml:"sitemap>loc"`
}

type News struct{
  Titles []string `xml:"url>news>title"`
  //Keywords []string `xml:"url>news>keywords"` //Keywords aren't currently available in the news tech sitemap
  PublicationDates []string `xml:"url>news>publication_date"`
  Locations []string `xml:"url>loc"`
}


func indexHandler (w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w,"<h1> Woah! Go is Neat!</h1>")
}

func newsRoutine(c chan News, Location string){
  defer wg.Done()
  var n News
  resp, _ := http.Get(Location)
  bytes, _ := ioutil.ReadAll(resp.Body)
  xml.Unmarshal(bytes,&n)
  resp.Body.Close()
  c <- n

}

func newsAggHandler (w http.ResponseWriter, r *http.Request) {

  var s Sitemapindex  
  //Make request using http.Get
  resp, _ := http.Get("https://www.thetimes.co.uk/sitemaps/sitemap.xml")
  bytes, _ := ioutil.ReadAll(resp.Body)
  xml.Unmarshal(bytes,&s)
  news_map := make(map[string]Newsmap)
  resp.Body.Close()
  queue := make(chan News,30)
  for _, Location := range s.Locations{
    wg.Add(1)
    go newsRoutine(queue, Location)
    // for idx, _ := range n.PublicationDates{
    //   news_map[n.Titles[idx]] = Newsmap{n.PublicationDates[idx],n.Locations[idx]}
    // }
  }
  wg.Wait()
  close(queue)

  for elem := range queue{
        for idx, _ := range elem.PublicationDates{
          news_map[elem.Titles[idx]] = Newsmap{elem.PublicationDates[idx],elem.Locations[idx]}
    }
  }
  
  p := NewsAggPage{Title: "Amazing News Aggregator", News: news_map}
  t, _ := template.ParseFiles("newsaggtemplate.html")
  t.Execute(w,p)
}

func main(){


  http.HandleFunc("/",indexHandler)
  http.HandleFunc("/agg/",newsAggHandler)
  http.ListenAndServe(":8000",nil)
}