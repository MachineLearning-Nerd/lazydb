package components

import (
	"context"
	"fmt"

	"github.com/MachineLearning-Nerd/lazydb/internal/db"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SchemaNode represents a node in the schema tree
type SchemaNode struct {
	Name       string
	Type       string // "schema", "tables", "views", "functions", "table", "view", "function"
	Schema     string
	Expanded   bool
	Children   []*SchemaNode
	ParentType string // For navigation
}

// SchemaTree manages the schema exploration tree
type SchemaTree struct {
	conn           db.Connection
	root           *SchemaNode
	flatList       []*SchemaNode // Flat list for navigation
	selectedIndex  int
	loading        bool
	err            error
	maxVisibleRows int
	scrollOffset   int
}

// NewSchemaTree creates a new schema tree
func NewSchemaTree(conn db.Connection) *SchemaTree {
	return &SchemaTree{
		conn:           conn,
		root:           &SchemaNode{Name: "Schemas", Type: "root", Expanded: true, Children: []*SchemaNode{}},
		flatList:       []*SchemaNode{},
		maxVisibleRows: 10,
	}
}

// LoadSchemas loads schemas from the database
func (st *SchemaTree) LoadSchemas(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		schemas, err := st.conn.ListSchemas(ctx)
		if err != nil {
			return SchemaErrorMsg{Err: err}
		}
		return SchemasLoadedMsg{Schemas: schemas}
	}
}

// LoadSchemaObjects loads tables, views, and functions for a schema
func (st *SchemaTree) LoadSchemaObjects(ctx context.Context, schema string) tea.Cmd {
	return func() tea.Msg {
		// Load tables
		tables, err := st.conn.ListTables(ctx, schema)
		if err != nil {
			return SchemaErrorMsg{Err: err}
		}

		// Load views
		views, err := st.conn.ListViews(ctx, schema)
		if err != nil {
			return SchemaErrorMsg{Err: err}
		}

		// Load functions
		functions, err := st.conn.ListFunctions(ctx, schema)
		if err != nil {
			return SchemaErrorMsg{Err: err}
		}

		return SchemaObjectsLoadedMsg{
			Schema:    schema,
			Tables:    tables,
			Views:     views,
			Functions: functions,
		}
	}
}

// LoadTableColumns loads columns for a table
func (st *SchemaTree) LoadTableColumns(ctx context.Context, schema, table string) tea.Cmd {
	return func() tea.Msg {
		columns, err := st.conn.GetTableColumns(ctx, schema, table)
		if err != nil {
			return SchemaErrorMsg{Err: err}
		}
		return TableColumnsLoadedMsg{
			Schema:  schema,
			Table:   table,
			Columns: columns,
		}
	}
}

// Toggle expands or collapses the selected node
func (st *SchemaTree) Toggle(ctx context.Context) tea.Cmd {
	if len(st.flatList) == 0 || st.selectedIndex >= len(st.flatList) {
		return nil
	}

	node := st.flatList[st.selectedIndex]
	node.Expanded = !node.Expanded

	// If expanding and no children, load them
	if node.Expanded && len(node.Children) == 0 {
		switch node.Type {
		case "schema":
			return st.LoadSchemaObjects(ctx, node.Name)
		case "tables":
			// Tables group - children already loaded
		case "views":
			// Views group - children already loaded
		case "functions":
			// Functions group - children already loaded
		case "table":
			return st.LoadTableColumns(ctx, node.Schema, node.Name)
		}
	}

	st.rebuildFlatList()
	return nil
}

// MoveDown moves selection down
func (st *SchemaTree) MoveDown() {
	if st.selectedIndex < len(st.flatList)-1 {
		st.selectedIndex++
		st.adjustScroll()
	}
}

// MoveUp moves selection up
func (st *SchemaTree) MoveUp() {
	if st.selectedIndex > 0 {
		st.selectedIndex--
		st.adjustScroll()
	}
}

// adjustScroll adjusts scroll offset to keep selection visible
func (st *SchemaTree) adjustScroll() {
	if st.selectedIndex < st.scrollOffset {
		st.scrollOffset = st.selectedIndex
	} else if st.selectedIndex >= st.scrollOffset+st.maxVisibleRows {
		st.scrollOffset = st.selectedIndex - st.maxVisibleRows + 1
	}
}

// GetSelected returns the currently selected node
func (st *SchemaTree) GetSelected() *SchemaNode {
	if len(st.flatList) == 0 || st.selectedIndex >= len(st.flatList) {
		return nil
	}
	return st.flatList[st.selectedIndex]
}

// HandleSchemasLoaded handles the schemas loaded message
func (st *SchemaTree) HandleSchemasLoaded(schemas []string) {
	st.root.Children = make([]*SchemaNode, len(schemas))
	for i, schema := range schemas {
		st.root.Children[i] = &SchemaNode{
			Name:     schema,
			Type:     "schema",
			Schema:   schema,
			Expanded: false,
			Children: []*SchemaNode{},
		}
	}
	st.rebuildFlatList()
}

// HandleSchemaObjectsLoaded handles the schema objects loaded message
func (st *SchemaTree) HandleSchemaObjectsLoaded(schema string, tables, views, functions []db.SchemaObject) {
	// Find the schema node
	var schemaNode *SchemaNode
	for _, node := range st.root.Children {
		if node.Name == schema {
			schemaNode = node
			break
		}
	}

	if schemaNode == nil {
		return
	}

	// Create category nodes
	schemaNode.Children = []*SchemaNode{}

	// Tables group
	if len(tables) > 0 {
		tablesNode := &SchemaNode{
			Name:       fmt.Sprintf("Tables (%d)", len(tables)),
			Type:       "tables",
			Schema:     schema,
			Expanded:   false,
			Children:   make([]*SchemaNode, len(tables)),
			ParentType: "schema",
		}
		for i, table := range tables {
			tablesNode.Children[i] = &SchemaNode{
				Name:       table.Name,
				Type:       "table",
				Schema:     schema,
				Expanded:   false,
				Children:   []*SchemaNode{},
				ParentType: "tables",
			}
		}
		schemaNode.Children = append(schemaNode.Children, tablesNode)
	}

	// Views group
	if len(views) > 0 {
		viewsNode := &SchemaNode{
			Name:       fmt.Sprintf("Views (%d)", len(views)),
			Type:       "views",
			Schema:     schema,
			Expanded:   false,
			Children:   make([]*SchemaNode, len(views)),
			ParentType: "schema",
		}
		for i, view := range views {
			viewsNode.Children[i] = &SchemaNode{
				Name:       view.Name,
				Type:       "view",
				Schema:     schema,
				Expanded:   false,
				Children:   []*SchemaNode{},
				ParentType: "views",
			}
		}
		schemaNode.Children = append(schemaNode.Children, viewsNode)
	}

	// Functions group
	if len(functions) > 0 {
		functionsNode := &SchemaNode{
			Name:       fmt.Sprintf("Functions (%d)", len(functions)),
			Type:       "functions",
			Schema:     schema,
			Expanded:   false,
			Children:   make([]*SchemaNode, len(functions)),
			ParentType: "schema",
		}
		for i, function := range functions {
			functionsNode.Children[i] = &SchemaNode{
				Name:       function.Name,
				Type:       "function",
				Schema:     schema,
				Expanded:   false,
				Children:   []*SchemaNode{},
				ParentType: "functions",
			}
		}
		schemaNode.Children = append(schemaNode.Children, functionsNode)
	}

	st.rebuildFlatList()
}

// HandleTableColumnsLoaded handles the table columns loaded message
func (st *SchemaTree) HandleTableColumnsLoaded(schema, table string, columns []db.TableColumn) {
	// Find the table node
	var tableNode *SchemaNode
	for _, schemaNode := range st.root.Children {
		if schemaNode.Schema == schema {
			for _, categoryNode := range schemaNode.Children {
				if categoryNode.Type == "tables" {
					for _, tblNode := range categoryNode.Children {
						if tblNode.Name == table {
							tableNode = tblNode
							break
						}
					}
				}
			}
		}
	}

	if tableNode == nil {
		return
	}

	// Create column nodes
	tableNode.Children = make([]*SchemaNode, len(columns))
	for i, col := range columns {
		nullable := ""
		if !col.Nullable {
			nullable = " NOT NULL"
		}
		defaultVal := ""
		if col.Default != "" {
			defaultVal = fmt.Sprintf(" DEFAULT %s", col.Default)
		}
		tableNode.Children[i] = &SchemaNode{
			Name:       fmt.Sprintf("%s: %s%s%s", col.Name, col.Type, nullable, defaultVal),
			Type:       "column",
			Schema:     schema,
			Expanded:   false,
			Children:   []*SchemaNode{},
			ParentType: "table",
		}
	}

	st.rebuildFlatList()
}

// rebuildFlatList rebuilds the flat list for navigation
func (st *SchemaTree) rebuildFlatList() {
	st.flatList = []*SchemaNode{}
	st.addNodeToFlatList(st.root, 0)
}

// addNodeToFlatList recursively adds nodes to flat list
func (st *SchemaTree) addNodeToFlatList(node *SchemaNode, depth int) {
	if node.Type != "root" {
		st.flatList = append(st.flatList, node)
	}

	if node.Expanded {
		for _, child := range node.Children {
			st.addNodeToFlatList(child, depth+1)
		}
	}
}

// View renders the schema tree
func (st *SchemaTree) View() string {
	if st.loading {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("6")).Render("Loading schema...")
	}

	if st.err != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(fmt.Sprintf("Error: %v", st.err))
	}

	if len(st.flatList) == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("No schemas loaded")
	}

	var output string
	visibleNodes := st.flatList[st.scrollOffset:]
	if len(visibleNodes) > st.maxVisibleRows {
		visibleNodes = visibleNodes[:st.maxVisibleRows]
	}

	for i, node := range visibleNodes {
		actualIndex := st.scrollOffset + i
		isSelected := actualIndex == st.selectedIndex

		line := st.renderNode(node, isSelected)
		output += line + "\n"
	}

	return output
}

// renderNode renders a single node
func (st *SchemaTree) renderNode(node *SchemaNode, selected bool) string {
	// Determine icon and style
	var icon, style string

	switch node.Type {
	case "schema":
		if node.Expanded {
			icon = "▼ 📂"
		} else {
			icon = "▶ 📂"
		}
		style = "6" // Cyan
	case "tables":
		if node.Expanded {
			icon = "  ▼ 📊"
		} else {
			icon = "  ▶ 📊"
		}
		style = "5" // Magenta
	case "views":
		if node.Expanded {
			icon = "  ▼ 👁"
		} else {
			icon = "  ▶ 👁"
		}
		style = "4" // Blue
	case "functions":
		if node.Expanded {
			icon = "  ▼ ⚙"
		} else {
			icon = "  ▶ ⚙"
		}
		style = "3" // Yellow
	case "table":
		if node.Expanded {
			icon = "    ▼ 📋"
		} else {
			icon = "    ▶ 📋"
		}
		style = "7" // White
	case "view":
		icon = "    > 👁"
		style = "7"
	case "function":
		icon = "    > ⚙"
		style = "7"
	case "column":
		icon = "      • "
		style = "8" // Gray
	default:
		icon = "  "
		style = "7"
	}

	text := fmt.Sprintf("%s %s", icon, node.Name)

	if selected {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("0")).
			Background(lipgloss.Color(style)).
			Bold(true).
			Render(text)
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(style)).
		Render(text)
}

// SetMaxVisibleRows sets the maximum visible rows
func (st *SchemaTree) SetMaxVisibleRows(rows int) {
	st.maxVisibleRows = rows
}

// Message types for schema operations

type SchemasLoadedMsg struct {
	Schemas []string
}

type SchemaObjectsLoadedMsg struct {
	Schema    string
	Tables    []db.SchemaObject
	Views     []db.SchemaObject
	Functions []db.SchemaObject
}

type TableColumnsLoadedMsg struct {
	Schema  string
	Table   string
	Columns []db.TableColumn
}

type SchemaErrorMsg struct {
	Err error
}
