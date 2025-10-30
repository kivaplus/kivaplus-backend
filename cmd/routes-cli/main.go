package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/kivaplus/kivaplus-backend/internal/shared/routes"
)

func main() {
	var (
		configPath = flag.String("config", "configs/routes.json", "Path to routes configuration file")
		command    = flag.String("cmd", "list", "Command: list, validate, add, remove, export")
		format     = flag.String("format", "table", "Output format: table, json, yaml")
		group      = flag.String("group", "", "Route group name")
		name       = flag.String("name", "", "Route name")
		method     = flag.String("method", "", "HTTP method")
		path       = flag.String("path", "", "Route path")
		lambda     = flag.String("lambda", "", "Lambda function name")
		public     = flag.Bool("public", false, "Is route public")
		roles      = flag.String("roles", "", "Comma-separated role IDs")
		resource   = flag.String("resource", "", "Resource name")
		action     = flag.String("action", "", "Action name")
		scope      = flag.String("scope", "global", "Permission scope")
	)
	flag.Parse()

	config, err := routes.LoadRoutesConfiguration(*configPath)
	if err != nil {
		fmt.Printf("Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	switch *command {
	case "list":
		listRoutes(config, *format, *group)
	case "validate":
		validateRoutes(config)
	case "add":
		addRoute(config, *configPath, *group, *name, *method, *path, *lambda, *public, *roles, *resource, *action, *scope)
	case "remove":
		removeRoute(config, *configPath, *name)
	case "export":
		exportRoutes(config, *format)
	case "generate":
		generateCode(config)
	default:
		fmt.Printf("Unknown command: %s\n", *command)
		printUsage()
		os.Exit(1)
	}
}

func listRoutes(config *routes.RoutesConfiguration, format, groupFilter string) {
	allRoutes := config.GetAllRoutes()

	if groupFilter != "" {
		var filteredRoutes []routes.RouteConfig
		for _, group := range config.Groups {
			if group.Name == groupFilter {
				filteredRoutes = group.Routes
				break
			}
		}
		allRoutes = filteredRoutes
	}

	switch format {
	case "json":
		data, _ := json.MarshalIndent(allRoutes, "", "  ")
		fmt.Println(string(data))
	case "table":
		printRoutesTable(allRoutes)
	default:
		fmt.Printf("Unsupported format: %s\n", format)
	}
}

func printRoutesTable(routes []routes.RouteConfig) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tMETHOD\tPATH\tPUBLIC\tLAMBDA\tROLES\tRESOURCE\tACTION\tSCOPE")
	fmt.Fprintln(w, "----\t------\t----\t------\t------\t-----\t--------\t------\t-----")

	for _, route := range routes {
		rolesStr := ""
		if len(route.Roles) > 0 {
			roleStrs := make([]string, len(route.Roles))
			for i, role := range route.Roles {
				roleStrs[i] = fmt.Sprintf("%d", role)
			}
			rolesStr = strings.Join(roleStrs, ",")
		}

		publicStr := "No"
		if route.Public {
			publicStr = "Yes"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			route.Name, route.Method, route.Path, publicStr, route.Lambda,
			rolesStr, route.Resource, route.Action, route.Scope)
	}
	w.Flush()
}

func validateRoutes(config *routes.RoutesConfiguration) {
	fmt.Println("Validating routes configuration...")

	// Basic validation
	allRoutes := config.GetAllRoutes()
	if len(allRoutes) == 0 {
		fmt.Println("❌ No routes defined")
		return
	}

	// Check for duplicate names
	nameMap := make(map[string]bool)
	duplicates := []string{}
	for _, route := range allRoutes {
		if nameMap[route.Name] {
			duplicates = append(duplicates, route.Name)
		}
		nameMap[route.Name] = true
	}

	if len(duplicates) > 0 {
		fmt.Printf("❌ Duplicate route names: %s\n", strings.Join(duplicates, ", "))
	}

	// Check for missing required fields
	errors := []string{}
	for _, route := range allRoutes {
		if route.Name == "" {
			errors = append(errors, "Route missing name")
		}
		if route.Method == "" {
			errors = append(errors, fmt.Sprintf("Route %s missing method", route.Name))
		}
		if route.Path == "" {
			errors = append(errors, fmt.Sprintf("Route %s missing path", route.Name))
		}
		if !route.Public && len(route.Roles) == 0 {
			errors = append(errors, fmt.Sprintf("Protected route %s has no roles", route.Name))
		}
	}

	if len(errors) > 0 {
		fmt.Println("❌ Validation errors:")
		for _, err := range errors {
			fmt.Printf("  - %s\n", err)
		}
		return
	}

	// Infrastructure validation
	infraAdapter := routes.NewInfrastructureAdapter(config)
	infraErrors := infraAdapter.ValidateConfiguration()
	if len(infraErrors) > 0 {
		fmt.Println("❌ Infrastructure validation errors:")
		for _, err := range infraErrors {
			fmt.Printf("  - %s\n", err)
		}
		return
	}

	fmt.Printf("✅ Configuration valid! Found %d routes across %d groups\n", len(allRoutes), len(config.Groups))

	// Summary
	publicCount := len(config.GetPublicRoutes())
	protectedCount := len(config.GetProtectedRoutes())
	fmt.Printf("📊 Summary: %d public routes, %d protected routes\n", publicCount, protectedCount)

	// Lambda distribution
	lambdaMappings := infraAdapter.GetLambdaMappings()
	fmt.Printf("🔧 Lambda functions: %d\n", len(lambdaMappings))
	for name, mapping := range lambdaMappings {
		fmt.Printf("  - %s: %d routes\n", name, len(mapping.Routes))
	}
}

func addRoute(config *routes.RoutesConfiguration, configPath, groupName, name, method, path, lambda string, public bool, rolesStr, resource, action, scope string) {
	if name == "" || method == "" || path == "" {
		fmt.Println("❌ Name, method, and path are required")
		return
	}

	// Parse roles
	var roleIDs []int
	if rolesStr != "" {
		roleStrs := strings.Split(rolesStr, ",")
		for _, roleStr := range roleStrs {
			var roleID int
			if _, err := fmt.Sscanf(strings.TrimSpace(roleStr), "%d", &roleID); err == nil {
				roleIDs = append(roleIDs, roleID)
			}
		}
	}

	// Create path pattern from path (simple conversion)
	pathPattern := "^" + strings.ReplaceAll(path, "{id}", "\\w+") + "$"
	pathPattern = strings.ReplaceAll(pathPattern, "{", "\\")
	pathPattern = strings.ReplaceAll(pathPattern, "}", "")

	newRoute := routes.RouteConfig{
		Name:        name,
		Method:      method,
		Path:        path,
		PathPattern: pathPattern,
		Public:      public,
		Roles:       roleIDs,
		Resource:    resource,
		Action:      action,
		Scope:       scope,
		Lambda:      lambda,
		Description: fmt.Sprintf("Auto-generated route for %s %s", method, path),
	}

	// Find or create group
	var targetGroup *routes.RouteGroup
	for i := range config.Groups {
		if config.Groups[i].Name == groupName {
			targetGroup = &config.Groups[i]
			break
		}
	}

	if targetGroup == nil {
		// Create new group
		newGroup := routes.RouteGroup{
			Name:        groupName,
			Description: fmt.Sprintf("Auto-generated group for %s", groupName),
			Routes:      []routes.RouteConfig{newRoute},
		}
		config.Groups = append(config.Groups, newGroup)
	} else {
		// Add to existing group
		targetGroup.Routes = append(targetGroup.Routes, newRoute)
	}

	// Save configuration
	if err := routes.SaveRoutesConfiguration(config, configPath); err != nil {
		fmt.Printf("❌ Error saving configuration: %v\n", err)
		return
	}

	fmt.Printf("✅ Added route %s to group %s\n", name, groupName)
}

func removeRoute(config *routes.RoutesConfiguration, configPath, name string) {
	found := false
	for groupIdx := range config.Groups {
		for routeIdx, route := range config.Groups[groupIdx].Routes {
			if route.Name == name {
				// Remove route
				config.Groups[groupIdx].Routes = append(
					config.Groups[groupIdx].Routes[:routeIdx],
					config.Groups[groupIdx].Routes[routeIdx+1:]...,
				)
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		fmt.Printf("❌ Route %s not found\n", name)
		return
	}

	// Save configuration
	if err := routes.SaveRoutesConfiguration(config, configPath); err != nil {
		fmt.Printf("❌ Error saving configuration: %v\n", err)
		return
	}

	fmt.Printf("✅ Removed route %s\n", name)
}

func exportRoutes(config *routes.RoutesConfiguration, format string) {
	switch format {
	case "json":
		data, _ := json.MarshalIndent(config, "", "  ")
		fmt.Println(string(data))
	default:
		fmt.Printf("Unsupported export format: %s\n", format)
	}
}

func generateCode(config *routes.RoutesConfiguration) {
	fmt.Println("🔧 Generating code from routes configuration...")

	// Generate authorization adapter usage example
	fmt.Println("\n// Authorization Service Usage:")
	fmt.Println("authAdapter := routes.NewAuthorizationAdapter(config)")
	fmt.Println("hasPermission, err := authAdapter.CheckPermission(method, path, claims)")

	// Generate infrastructure adapter usage example
	fmt.Println("\n// Infrastructure Setup Usage:")
	fmt.Println("infraAdapter := routes.NewInfrastructureAdapter(config)")
	fmt.Println("apiRoutes := infraAdapter.GetAPIRoutes(lambdaFunctions)")

	fmt.Println("\n✅ Code generation examples printed above")
}

func printUsage() {
	fmt.Println("Routes CLI - Manage API routes configuration")
	fmt.Println("\nUsage:")
	fmt.Println("  routes-cli -cmd=list                           # List all routes")
	fmt.Println("  routes-cli -cmd=validate                       # Validate configuration")
	fmt.Println("  routes-cli -cmd=add -group=users -name=create_user -method=POST -path=/usuarios -lambda=users")
	fmt.Println("  routes-cli -cmd=remove -name=create_user       # Remove route")
	fmt.Println("  routes-cli -cmd=export -format=json           # Export configuration")
	fmt.Println("  routes-cli -cmd=generate                       # Generate code examples")
	fmt.Println("\nFlags:")
	flag.PrintDefaults()
}
