package server

// Tool Categories for selective loading and filtering
const (
	CategorySchema        = "schema"        // Table structure, DDL, definitions
	CategoryPerformance   = "performance"   // EXPLAIN, indexes, sizes
	CategoryRelationships = "relationships" // FK, constraints, dependencies
	CategoryStatistics    = "statistics"    // Column/table stats
	CategoryDiscovery     = "discovery"     // Search, sample, sequences
	CategoryTriggers      = "triggers"      // Trigger definitions and listings
	CategoryOptimization  = "optimization"  // Query optimization tools
	CategoryMeta          = "meta"          // Tool discovery and search
	CategoryCompound      = "compound"      // Multi-operation tools
)

// CategoryTools maps categories to their tool names
var CategoryTools = map[string][]string{
	CategorySchema: {
		"list_all_tables",
		"get_table_schema",
		"search_tables",
		"get_table_ddl",
		"get_view_definition",
		"get_function_definition",
	},
	CategoryPerformance: {
		"explain_query",
		"get_table_indexes",
		"get_table_size",
	},
	CategoryRelationships: {
		"get_foreign_keys",
		"get_table_constraints",
		"get_table_dependencies",
		"get_table_references",
	},
	CategoryStatistics: {
		"get_column_stats",
		"get_table_stats",
	},
	CategoryDiscovery: {
		"get_sample_data",
		"get_table_count",
		"list_sequences",
		"list_materialized_views",
	},
	CategoryTriggers: {
		"get_table_triggers",
		"get_trigger_definition",
	},
	CategoryOptimization: {
		"analyze_query_performance",
		"get_query_optimization_context",
		"compare_query_performance",
	},
	CategoryMeta: {
		"search_lazydb_tools",      // Key tool for dynamic tool discovery
		"lazydb_enable_category",   // Enable a category at runtime
		"lazydb_disable_category",  // Disable a category at runtime
		"lazydb_list_categories",   // List all categories with status
		"lazydb_reset_session",     // Reset session to a preset
	},
	CategoryCompound: {
		"analyze_table_comprehensive",
	},
}

// CategoryDescriptions provides human-readable descriptions for categories
var CategoryDescriptions = map[string]string{
	CategorySchema:        "Table structure and definitions (DDL, views, functions)",
	CategoryPerformance:   "Performance analysis (EXPLAIN, indexes, sizes)",
	CategoryRelationships: "Foreign keys, constraints, and dependencies",
	CategoryStatistics:    "Column and table statistics",
	CategoryDiscovery:     "Data discovery (samples, counts, sequences)",
	CategoryTriggers:      "Database triggers and events",
	CategoryOptimization:  "Query optimization and analysis",
	CategoryMeta:          "Tool discovery and search (find other tools dynamically)",
	CategoryCompound:      "Multi-operation tools (combine multiple queries)",
}

// AllCategories returns all available categories
func AllCategories() []string {
	return []string{
		CategorySchema,
		CategoryPerformance,
		CategoryRelationships,
		CategoryStatistics,
		CategoryDiscovery,
		CategoryTriggers,
		CategoryOptimization,
		CategoryMeta,
		CategoryCompound,
	}
}

// GetCategoryTools returns tool names for a given category
func GetCategoryTools(category string) []string {
	if tools, ok := CategoryTools[category]; ok {
		return tools
	}
	return nil
}

// GetToolCategory returns the category for a given tool name
func GetToolCategory(toolName string) string {
	for category, tools := range CategoryTools {
		for _, tool := range tools {
			if tool == toolName {
				return category
			}
		}
	}
	return ""
}

// CategoryPresets provides common category combinations
var CategoryPresets = map[string][]string{
	"minimal": {
		CategorySchema,
		CategoryMeta, // Tool discovery - enables finding other tools dynamically
	},
	"standard": {
		CategorySchema,
		CategoryDiscovery,
		CategoryRelationships,
		CategoryMeta,
	},
	"performance": {
		CategoryPerformance,
		CategoryStatistics,
		CategoryOptimization,
		CategoryMeta,
	},
	"full": {
		CategorySchema,
		CategoryPerformance,
		CategoryRelationships,
		CategoryStatistics,
		CategoryDiscovery,
		CategoryTriggers,
		CategoryOptimization,
		CategoryMeta,
		CategoryCompound,
	},
}

// GetPresetCategories returns categories for a preset name
func GetPresetCategories(preset string) []string {
	if categories, ok := CategoryPresets[preset]; ok {
		return categories
	}
	return nil
}
