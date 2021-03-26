package db

var DefaultAdapterMiddleware = &AdapterMiddleware{}

type AdapterMiddleware struct {
	creates    []*func(scope *AdapterMiddlewareScope)
	updates    []*func(scope *AdapterMiddlewareScope)
	deletes    []*func(scope *AdapterMiddlewareScope)
	queries    []*func(scope *AdapterMiddlewareScope)
	processors []*AdapterMiddlewareProcessor
}

func (m *AdapterMiddleware) clone() *AdapterMiddleware {
	return &AdapterMiddleware{
		creates:    m.creates,
		updates:    m.updates,
		deletes:    m.deletes,
		queries:    m.queries,
		processors: m.processors,
	}
}
func (m *AdapterMiddleware) NewScope(model IModel, v interface{}) *AdapterMiddlewareScope {
	return &AdapterMiddlewareScope{model: model, InputValue: v}
}

func (m *AdapterMiddleware) Create() *AdapterMiddlewareProcessor {
	return &AdapterMiddlewareProcessor{kind: "create", parent: m}
}

func (m *AdapterMiddleware) Update() *AdapterMiddlewareProcessor {
	return &AdapterMiddlewareProcessor{kind: "update", parent: m}
}

func (m *AdapterMiddleware) Delete() *AdapterMiddlewareProcessor {
	return &AdapterMiddlewareProcessor{kind: "delete", parent: m}
}

func (m *AdapterMiddleware) Query() *AdapterMiddlewareProcessor {
	return &AdapterMiddlewareProcessor{kind: "query", parent: m}
}

func (m *AdapterMiddleware) reorder() {
	var creates, updates, deletes, queries []*AdapterMiddlewareProcessor

	for _, processor := range m.processors {
		if processor.name != "" {
			switch processor.kind {
			case "create":
				creates = append(creates, processor)
			case "update":
				updates = append(updates, processor)
			case "delete":
				deletes = append(deletes, processor)
			case "query":
				queries = append(queries, processor)
			}
		}
	}

	m.creates = sortProcessors(creates)
	m.updates = sortProcessors(updates)
	m.deletes = sortProcessors(deletes)
	m.queries = sortProcessors(queries)
}

type AdapterMiddlewareProcessor struct {
	name      string                               // current callback's name
	before    string                               // register current callback before a callback
	after     string                               // register current callback after a callback
	replace   bool                                 // replace callbacks with same name
	remove    bool                                 // delete callbacks with same name
	kind      string                               // callback type: create, update, delete, query, row_query
	processor *func(scope *AdapterMiddlewareScope) // callback handler
	parent    *AdapterMiddleware
}

func (p *AdapterMiddlewareProcessor) After(callbackName string) *AdapterMiddlewareProcessor {
	p.after = callbackName
	return p
}

func (p *AdapterMiddlewareProcessor) Before(callbackName string) *AdapterMiddlewareProcessor {
	p.before = callbackName
	return p
}

func (p *AdapterMiddlewareProcessor) Register(callbackName string, callback func(scope *AdapterMiddlewareScope)) {
	p.name = callbackName
	p.processor = &callback
	p.parent.processors = append(p.parent.processors, p)
	p.parent.reorder()
}

func (p *AdapterMiddlewareProcessor) Remove(callbackName string) {
	p.name = callbackName
	p.remove = true
	p.parent.processors = append(p.parent.processors, p)
	p.parent.reorder()
}
func (p *AdapterMiddlewareProcessor) Replace(callbackName string, callback func(scope *AdapterMiddlewareScope)) {
	p.name = callbackName
	p.processor = &callback
	p.replace = true
	p.parent.processors = append(p.parent.processors, p)
	p.parent.reorder()
}
func (p *AdapterMiddlewareProcessor) Get(callbackName string) (callback func(scope *AdapterMiddlewareScope)) {
	for _, p := range p.parent.processors {
		if p.name == callbackName && p.kind == p.kind {
			if p.remove {
				callback = nil
			} else {
				callback = *p.processor
			}
		}
	}
	return
}
func getRIndex(strs []string, str string) int {
	for i := len(strs) - 1; i >= 0; i-- {
		if strs[i] == str {
			return i
		}
	}
	return -1
}

func sortProcessors(ps []*AdapterMiddlewareProcessor) []*func(scope *AdapterMiddlewareScope) {
	var (
		allNames, sortedNames          []string
		sortAdapterMiddlewareProcessor func(c *AdapterMiddlewareProcessor)
	)

	for _, p := range ps {
		allNames = append(allNames, p.name)
	}

	sortAdapterMiddlewareProcessor = func(c *AdapterMiddlewareProcessor) {
		if getRIndex(sortedNames, c.name) == -1 { // if not sorted
			if c.before != "" { // if defined before callback
				if index := getRIndex(sortedNames, c.before); index != -1 {
					// if before callback already sorted, append current callback just after it
					sortedNames = append(sortedNames[:index], append([]string{c.name}, sortedNames[index:]...)...)
				} else if index := getRIndex(allNames, c.before); index != -1 {
					// if before callback exists but haven't sorted, append current callback to last
					sortedNames = append(sortedNames, c.name)
					sortAdapterMiddlewareProcessor(ps[index])
				}
			}

			if c.after != "" { // if defined after callback
				if index := getRIndex(sortedNames, c.after); index != -1 {
					// if after callback already sorted, append current callback just before it
					sortedNames = append(sortedNames[:index+1], append([]string{c.name}, sortedNames[index+1:]...)...)
				} else if index := getRIndex(allNames, c.after); index != -1 {
					// if after callback exists but haven't sorted
					p := ps[index]
					// set after callback's before callback to current callback
					if p.before == "" {
						p.before = c.name
					}
					sortAdapterMiddlewareProcessor(p)
				}
			}

			// if current callback haven't been sorted, append it to last
			if getRIndex(sortedNames, c.name) == -1 {
				sortedNames = append(sortedNames, c.name)
			}
		}
	}

	for _, p := range ps {
		sortAdapterMiddlewareProcessor(p)
	}

	var sortedFuncs []*func(scope *AdapterMiddlewareScope)
	for _, name := range sortedNames {
		if index := getRIndex(allNames, name); !ps[index].remove {
			sortedFuncs = append(sortedFuncs, ps[index].processor)
		}
	}

	return sortedFuncs
}
