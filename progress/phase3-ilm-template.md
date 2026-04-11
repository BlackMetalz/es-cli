# Phase 3: ILM Policies + Index Templates

## Status: Complete

## Features Implemented

### ILM Policies View (`:ilm` / `:ilm-policy`)
- Table: name, version, delete after
- **Sort**: Shift+N by name, toggle ASC/DESC
- **Search**: `/` filter by policy name
- **Hide system/managed**: `a` toggles, default hides `.` prefixed + `_meta.managed: true` policies
- **Create** (`n`): simple form — Name + Delete After (e.g. "30d"). Auto-generates hot phase with rollover defaults + delete phase
- **Edit** (`e`): fetches existing policy, pre-fills Delete After, name is read-only. Flash shows "updated" vs "created"
- **Delete** (`d`): confirm overlay with proper text "delete ILM policy 'name'?"
- **View detail** (Enter): JSON viewer via `GET /_ilm/policy/{name}`
- Selected row highlight on name column

### Index Templates View (`:template` / `:templates` / `:index-template`)
- Table: name, index_patterns, shards, replicas, ilm_policy (priority removed, auto-set to 100)
- **Sort**: Shift+N by name
- **Search**: `/` filter by template name or index pattern
- **Hide system/managed**: `a` toggles, default hides `.` prefixed + `_meta.managed: true` templates
- **Create** (`n`): 2-pane form — left: Name, Patterns, Shards, Replicas, ILM Policy; right: live pretty-printed JSON preview
- **Edit** (`e`): fetches existing template, pre-fills all fields, name read-only. Flash shows "updated" vs "created"
- **Delete** (`d`): confirm overlay with "delete index template 'name'?"
- **View detail** (Enter): JSON viewer via `GET /_index_template/{name}`
- Selected row highlight on name column

### ILM Policy Autocomplete in Template Form
- When creating/editing a template, ILM Policy field shows fuzzy suggestions
- Only shows user ILM policies (no `.` prefix, no managed)
- Up to 5 suggestions shown below the field as you type
- **Tab** auto-completes when exactly 1 match
- Fetched async when opening the form

### Duplicate Detection in Template Form
- **Template name**: warns "Template 'x' already exists" in real-time (yellow ⚠)
- **Pattern overlap**: warns "Pattern 'x' overlaps with template 'y' (z)" using prefix matching
- **Enter blocked** while warnings exist — must fix before submitting
- Warnings auto-clear when user fixes the issue
- Checks all existing templates (including system/managed) for overlaps

### Error Handling
- Operation errors (create/edit/delete failures) show as **modal popup** instead of flash
- Error text word-wrapped to fit screen
- Press **Enter** or **Esc** to dismiss — no auto-disappear timer
- Confirm overlays show proper resource type: "ILM policy" / "index template" / "index"

### Generic JSON Viewer
- New reusable `jsonview` package — accepts any fetch function, renders pretty-printed JSON in scrollable viewport
- Used by ILM detail, template detail
- Esc goes back to list view

### ES Client Additions
- `ilm.go`: `ILMPolicy` struct (with `Managed` flag), `ListILMPolicies()`, `GetILMPolicy()`, `CreateILMPolicy()`, `DeleteILMPolicy()`
- `template.go`: `IndexTemplate` struct (with `Managed` flag), `ListIndexTemplates()`, `GetIndexTemplate()`, `CreateIndexTemplate()`, `DeleteIndexTemplate()`

### App.go Changes
- Router: registered `ilm` and `template` commands with aliases
- **Context-aware `n` key**: dispatches based on `currentView().Name()` — Indices→createindex, ILM→createilm, Templates→createtemplate
- New overlay types: `overlayCreateILM`, `overlayCreateTemplate`, `overlayError`
- Edit flows: `edit_ilm` and `edit_template` pending actions fetch data, open pre-filled forms
- Error popup overlay: modal with word-wrapped error, Enter/Esc to dismiss

## New Files
```
internal/es/ilm.go + ilm_test.go
internal/es/template.go + template_test.go

internal/tui/views/ilm/
  model.go, keybindings.go, sort.go, render.go, filter.go, model_test.go

internal/tui/views/template/
  model.go, keybindings.go, sort.go, render.go, filter.go, model_test.go

internal/tui/views/jsonview/jsonview.go

internal/tui/components/createilm/createilm.go + createilm_test.go
internal/tui/components/createtemplate/createtemplate.go + createtemplate_test.go
```

## Test Coverage
- All 17+ packages passing
- ES client: ILM + template CRUD tests with httptest mocks
- ILM view: 11 tests (New, Load, Sort, Delete, Edit, Refresh, Error, View, HelpGroups, SetSize, IsInputMode)
- Template view: 12 tests
- Create ILM form: 8 tests (New, NewEdit, Tab, Submit valid/empty, Cancel, BuildJSON, View)
- Create template form: 7 tests (New, Tab, Submit valid/empty name/patterns, Cancel, BuildJSON)
- Command router: ILM + template autocomplete tests
