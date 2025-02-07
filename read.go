package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-shiori/go-readability"
)

func main() {
	urls := []string{
		"https://www.businessinsider.com/mark-zuckerberg-meta-ai-replace-engineers-coders-joe-rogan-podcast-2025-1",
		"https://www.latent.space/p/2025-papers",
		"https://thehustle.co/originals/how-corn-syrup-took-over-america",
		"https://www.barrons.com/articles/jpmorgan-back-to-office-mandate-union-4206af78",
		"https://www.404media.co/ceo-of-ai-music-company-says-people-dont-like-making-music/",
		"https://www.newyorker.com/magazine/2025/01/20/lorne-michaels-profile",
		"https://www.theguardian.com/uk-news/2025/jan/13/most-violent-or-sexual-offences-went-unsolved-in-uk-hotspots-last-year",
		"https://medium.com/@dennyluan/the-great-firewall-is-finally-broken-deb82428481a",
		"https://www.dailymail.co.uk/health/article-14280007/researchers-warn-chronic-wasting-disease-zombie-deer.html",
		"https://www.githubstatus.com/incidents/qd96yfgvmcf9",
		"https://arstechnica.com/tech-policy/2025/01/judge-ends-mans-11-year-quest-to-dig-up-landfill-and-recover-765m-in-bitcoin/",
		"https://www.theguardian.com/lifeandstyle/2024/nov/17/bank-of-mum-and-dad-why-we-all-now-live-in-an-inheritocracy",
		"https://gothamist.com/news/43k-fewer-drivers-on-manhattan-roads-after-congestion-pricing-turned-on-mta-says",
		"https://jaymartin.substack.com/p/has-canada-become-a-jamaican-bobsled",
		"https://www.opennet.ru/opennews/art.shtml?num=62551",
		"https://github.com/campsite/campsite",
		"https://www.opennet.ru/opennews/art.shtml?num=62552",
		"https://habr.com/ru/articles/873046/?utm_campaign=873046&utm_source=habrahabr&utm_medium=rss",
		"https://habr.com/ru/companies/wirenboard/articles/873394/?utm_campaign=873394&utm_source=habrahabr&utm_medium=rss",
		"https://www.opennet.ru/opennews/art.shtml?num=62553",
		"https://github.com/openzfs/zfs/releases/tag/zfs-2.3.0",
		"https://www.rednote.pro/",
		"https://habr.com/ru/articles/872890/?utm_campaign=872890&utm_source=habrahabr&utm_medium=rss",
		"https://www.gallup.com/workplace/654911/employee-engagement-sinks-year-low.aspx",
		"https://habr.com/ru/companies/ruvds/articles/872378/?utm_campaign=872378&utm_source=habrahabr&utm_medium=rss",
		"https://habr.com/ru/articles/873490/?utm_campaign=873490&utm_source=habrahabr&utm_medium=rss",
		"https://habr.com/ru/articles/872876/?utm_campaign=872876&utm_source=habrahabr&utm_medium=rss",
		"https://habr.com/ru/articles/873520/?utm_campaign=873520&utm_source=habrahabr&utm_medium=rss",
		"https://www.bbc.com/news/articles/cg52543v6rmo",
		"https://db-engines.com/en/blog_post/109",
		"https://restofworld.org/2025/singapore-ai-eldercare-tools/",
		"https://www.ornl.gov/news/plant-co2-uptake-rises-nearly-one-third-new-global-estimates",
		"https://habr.com/ru/companies/ruvds/articles/871288/?utm_campaign=871288&utm_source=habrahabr&utm_medium=rss",
		"https://fsfe.org/news/nl/nl-202501.en.html",
		"https://www.axios.com/2025/01/14/workers-job-satisfaction-gallup",
		"https://www.ycombinator.com/companies/craniometrix/jobs/5Ucqf0Q-founding-full-stack-engineer-cto-track",
		"https://proton.me/blog/2024-lifetime-fundraiser-results",
		"https://cleanlabelproject.org/wp-content/uploads/CleanLabelProject_ProteinStudyWhitepaper_010625.pdf",
		"https://insideevs.com/news/745119/tesla-sales-europe-2024/",
		"https://en.wikipedia.org/wiki/Jeppson%27s_Mal%C3%B6rt",
		"https://collegetowns.substack.com/p/making-an-intersection-unsafe-for",
		"https://studenttheses.uu.nl/bitstream/handle/20.500.12932/47209/Thesis_Final.pdf?sequence=1&isAllowed=y",
		"https://newsletter.goodtechthings.com/p/why-are-tech-people-suddenly-so-into",
		"https://arstechnica.com/gadgets/2025/01/allstate-sued-for-allegedly-tracking-drivers-behavior-through-third-party-apps/",
		"https://www.cnbc.com/2025/01/14/meta-targeting-lowest-performing-employees-in-latest-round-of-layoffs.html",
		"https://www.whitehouse.gov/briefing-room/presidential-actions/2025/01/14/executive-order-on-advancing-united-states-leadership-in-artificial-intelligence-infrastructure/",
		"https://twitter.com/MrBeast/status/1879224239485808811",
		"https://levels.fyi/heatmap/europe/",
		"https://news.ycombinator.com/item?id=42701745",
		"https://www.opennet.ru/opennews/art.shtml?num=62554",
	}

	for _, url := range urls {
		article, err := readability.FromURL(url, 10*time.Second)
		if err != nil {
			log.Printf("failed to parse %s, %v\n", url, err)
			continue
		}

		title := strings.ReplaceAll(article.Title, ".:,!?", "")
		title = strings.ReplaceAll(title, " ", "_")

		filename := fmt.Sprintf("%s.html", title)

		err = os.WriteFile("./data/"+filename, []byte(article.Content), 0644)

		fmt.Println(article.Title)
	}
}
