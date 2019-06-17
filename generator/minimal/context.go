package minimal

type APIContext struct {
	contexts    *APIContexts
	filename    string
	Models      []*Model
	Services    []*Service
	TwirpPrefix string
}

type APIContexts struct {
	modelLookup map[string]*Model
	contexts    []*APIContext
	twirpPrefix string
}

func NewAPIContexts(twirpPrefix string) *APIContexts {
	return &APIContexts{
		modelLookup: make(map[string]*Model),
		twirpPrefix: twirpPrefix,
	}
}

func (cs *APIContexts) addContext(filename string) {
	cs.contexts = append(cs.contexts, &APIContext{contexts: cs, filename: filename, TwirpPrefix: cs.twirpPrefix})
}

func (cs *APIContexts) AddModel(m *Model) {
	ctx := cs.contexts[len(cs.contexts)-1]
	ctx.Models = append(ctx.Models, m)
	cs.modelLookup[m.Name] = m
}

func (cs *APIContexts) AddService(s *Service) {
	ctx := cs.contexts[len(cs.contexts)-1]
	ctx.Services = append(ctx.Services, s)
}
