package detect

import (
	"os"
	"path/filepath"

	"github.com/ywnaing/term/internal/config"
)

type ProjectKind string

const (
	Generic         ProjectKind = "generic"
	Node            ProjectKind = "node"
	Maven           ProjectKind = "maven"
	Gradle          ProjectKind = "gradle"
	DockerCompose   ProjectKind = "docker-compose"
	Frontend        ProjectKind = "frontend"
	DotNet          ProjectKind = "dotnet"
	Go              ProjectKind = "go"
	SpringFullstack ProjectKind = "spring-fullstack"
)

func Detect(dir string) []ProjectKind {
	var kinds []ProjectKind
	if exists(dir, "package.json") {
		kinds = append(kinds, Node)
	}
	if exists(dir, "pom.xml") {
		kinds = append(kinds, Maven)
	}
	if exists(dir, "build.gradle") || exists(dir, "build.gradle.kts") {
		kinds = append(kinds, Gradle)
	}
	if hasCompose(dir) {
		kinds = append(kinds, DockerCompose)
	}
	if exists(dir, "frontend", "package.json") {
		kinds = append(kinds, Frontend)
	}
	if hasExt(dir, ".csproj") {
		kinds = append(kinds, DotNet)
	}
	if exists(dir, "go.mod") {
		kinds = append(kinds, Go)
	}
	if contains(kinds, Maven) && contains(kinds, Frontend) && contains(kinds, DockerCompose) {
		kinds = append(kinds, SpringFullstack)
	}
	if len(kinds) == 0 {
		kinds = append(kinds, Generic)
	}
	return kinds
}

func DefaultConfig(dir string) config.TermConfig {
	project := config.ProjectNameFromDir(dir)
	kinds := Detect(dir)
	if contains(kinds, SpringFullstack) {
		return config.TermConfig{Project: project, Shortcuts: map[string]config.Shortcut{
			"setup":    {Description: "Install dependencies", Steps: steps("./mvnw clean install -DskipTests", "cd frontend && npm install")},
			"dev":      {Description: "Start local development", Parallel: true, Steps: []config.Step{{Name: "database", Command: "docker compose up db"}, {Name: "backend", Command: "./mvnw spring-boot:run"}, {Name: "frontend", Command: "cd frontend && npm run dev"}}},
			"test":     {Description: "Run backend and frontend tests", Steps: steps("./mvnw test", "cd frontend && npm test")},
			"build":    {Description: "Build backend and frontend", Steps: steps("./mvnw clean package", "cd frontend && npm run build")},
			"db":       {Description: "Open local PostgreSQL shell", Steps: steps("docker compose exec db psql -U postgres")},
			"reset-db": {Description: "Reset local database", Danger: "high", Confirm: true, Steps: steps("docker compose down -v", "docker compose up -d db")},
		}}
	}
	if contains(kinds, Node) {
		return config.TermConfig{Project: project, Shortcuts: map[string]config.Shortcut{
			"dev":   {Description: "Start development server", Steps: steps("npm run dev")},
			"test":  {Description: "Run tests", Steps: steps("npm test")},
			"build": {Description: "Build project", Steps: steps("npm run build")},
		}}
	}
	if contains(kinds, Maven) {
		return config.TermConfig{Project: project, Shortcuts: map[string]config.Shortcut{
			"dev":   {Description: "Start Spring Boot application", Steps: steps("./mvnw spring-boot:run")},
			"test":  {Description: "Run tests", Steps: steps("./mvnw test")},
			"build": {Description: "Build jar", Steps: steps("./mvnw clean package")},
		}}
	}
	return config.TermConfig{Project: project, Shortcuts: map[string]config.Shortcut{
		"dev":  {Description: "Start local development", Steps: steps(`echo "Configure your dev command in .term.yml"`)},
		"test": {Description: "Run tests", Steps: steps(`echo "Configure your test command in .term.yml"`)},
	}}
}

func steps(commands ...string) []config.Step {
	out := make([]config.Step, 0, len(commands))
	for _, command := range commands {
		out = append(out, config.Step{Command: command})
	}
	return out
}

func exists(dir string, parts ...string) bool {
	_, err := os.Stat(filepath.Join(append([]string{dir}, parts...)...))
	return err == nil
}

func hasCompose(dir string) bool {
	for _, name := range []string{"docker-compose.yml", "docker-compose.yaml", "compose.yml", "compose.yaml"} {
		if exists(dir, name) {
			return true
		}
	}
	return false
}

func hasExt(dir, ext string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ext {
			return true
		}
	}
	return false
}

func contains(kinds []ProjectKind, want ProjectKind) bool {
	for _, kind := range kinds {
		if kind == want {
			return true
		}
	}
	return false
}
