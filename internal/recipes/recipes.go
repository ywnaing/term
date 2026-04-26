package recipes

import (
	_ "embed"
	"encoding/json"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/lithammer/fuzzysearch/fuzzy"
)

//go:embed recipes.json
var embedded []byte

type Recipe struct {
	ID          string              `json:"id"`
	Title       string              `json:"title"`
	Category    string              `json:"category"`
	Keywords    []string            `json:"keywords"`
	Description string              `json:"description"`
	Commands    map[string][]string `json:"commands"`
	Notes       []string            `json:"notes"`
}

func Load() ([]Recipe, error) {
	var recipes []Recipe
	err := json.Unmarshal(embedded, &recipes)
	return recipes, err
}

func Search(query string, all []Recipe) []Recipe {
	q := strings.ToLower(strings.TrimSpace(query))
	type scored struct {
		recipe Recipe
		score  int
	}
	var hits []scored
	for _, recipe := range all {
		hay := strings.ToLower(recipe.Title + " " + recipe.Category + " " + recipe.Description + " " + strings.Join(recipe.Keywords, " "))
		score := 0
		if strings.Contains(hay, q) {
			score += 100
		}
		for _, word := range strings.Fields(q) {
			if strings.Contains(hay, word) {
				score += 10
			}
		}
		if fuzzy.Match(q, hay) {
			score += 20
		}
		if score > 0 {
			hits = append(hits, scored{recipe: recipe, score: score})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].score > hits[j].score })
	out := make([]Recipe, 0, len(hits))
	for _, hit := range hits {
		out = append(out, hit.recipe)
	}
	return out
}

func ExtractPort(text string) string {
	re := regexp.MustCompile(`\b([1-9][0-9]{1,4})\b`)
	for _, match := range re.FindAllStringSubmatch(text, -1) {
		return match[1]
	}
	return ""
}

func ReplaceVars(commands []string, port string) []string {
	out := make([]string, len(commands))
	for i, command := range commands {
		if port != "" {
			command = strings.ReplaceAll(command, "<PORT>", port)
		}
		out[i] = command
	}
	return out
}

func CurrentOSGroup() string {
	switch runtime.GOOS {
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	case "windows":
		return "Windows"
	default:
		return runtime.GOOS
	}
}
