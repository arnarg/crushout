package rules

// Default is the built-in rule set.
var Default = map[string]*Rule{
	// ── Non-mutable ──────────────────────────────────────────
	"ls":       {Default: Allow},
	"cat":      {Default: Allow},
	"head":     {Default: Allow},
	"tail":     {Default: Allow},
	"less":     {Default: Allow},
	"bat":      {Default: Allow},
	"file":     {Default: Allow},
	"stat":     {Default: Allow},
	"wc":       {Default: Allow},
	"grep":     {Default: Allow},
	"rg":       {Default: Allow},
	"ag":       {Default: Allow},
	"sort":     {Default: Allow},
	"uniq":     {Default: Allow},
	"cut":      {Default: Allow},
	"tr":       {Default: Allow},
	"diff":     {Default: Allow},
	"comm":     {Default: Allow},
	"tac":      {Default: Allow},
	"rev":      {Default: Allow},
	"paste":    {Default: Allow},
	"column":   {Default: Allow},
	"expand":   {Default: Allow},
	"tree":     {Default: Allow},
	"pwd":      {Default: Allow},
	"whoami":   {Default: Allow},
	"hostname": {Default: Allow},
	"uname":    {Default: Allow},
	"date":     {Default: Allow},
	"printenv": {Default: Allow},
	"which":    {Default: Allow},
	"echo":     {Default: Allow},
	"printf":   {Default: Allow},
	"test":     {Default: Allow},
	"true":     {Default: Allow},
	"false":    {Default: Allow},
	"basename": {Default: Allow},
	"dirname":  {Default: Allow},
	"realpath": {Default: Allow},
	"readlink": {Default: Allow},
	"ps":       {Default: Allow},
	"df":       {Default: Allow},
	"du":       {Default: Allow},
	"free":     {Default: Allow},
	"ss":       {Default: Allow},
	"dig":      {Default: Allow},
	"ping":     {Default: Allow},
	"seq":      {Default: Allow},
	"sleep":    {Default: Allow},
	"bc":       {Default: Allow},
	"expr":     {Default: Allow},

	// Data processing
	"jq":        {Default: Allow},
	"base64":    {Default: Allow},
	"md5sum":    {Default: Allow},
	"sha1sum":   {Default: Allow},
	"sha256sum": {Default: Allow},
	"sha512sum": {Default: Allow},
	"b3sum":     {Default: Allow},
	"hexdump":   {Default: Allow},
	"xxd":       {Default: Allow},
	"od":        {Default: Allow},

	// Flag-dependent
	"yq": {
		Default:     Allow,
		PromptFlags: []string{"-i", "--in-place"},
	},
	"sed": {
		Default:     Allow,
		PromptFlags: []string{"-i", "--in-place"},
	},
	"find": {
		Default:     Allow,
		PromptFlags: []string{"-exec", "-execdir", "-ok", "-okdir", "-delete", "-fprint", "-fls"},
	},

	// ── Nested subcommands ──────────────────────────────────
	"git": {
		PromptFlags: []string{"-C", "--work-tree", "--git-dir"},
		Subcommands: map[string]*Rule{
			"status":        {Default: Allow},
			"diff":          {Default: Allow},
			"log":           {Default: Allow},
			"show":          {Default: Allow},
			"blame":         {Default: Allow},
			"shortlog":      {Default: Allow},
			"reflog":        {Default: Allow},
			"whatchanged":   {Default: Allow},
			"ls-files":      {Default: Allow},
			"ls-tree":       {Default: Allow},
			"ls-remote":     {Default: Allow},
			"describe":      {Default: Allow},
			"rev-parse":     {Default: Allow},
			"name-rev":      {Default: Allow},
			"merge-base":    {Default: Allow},
			"cherry":        {Default: Allow},
			"count-objects": {Default: Allow},
			"fsck":          {Default: Allow},
			"check-ignore":  {Default: Allow},
			"help":          {Default: Allow},
			"version":       {Default: Allow},
			"range-diff":    {Default: Allow},
			"var":           {Default: Allow},
			"remote": {
				Subcommands: map[string]*Rule{
					"-v":        {Default: Allow},
					"--verbose": {Default: Allow},
					"show":      {Default: Allow},
					"list":      {Default: Allow},
					"get-url":   {Default: Allow},
				},
			},
			"branch": {
				PromptFlags: []string{"-d", "-D", "--delete", "-m", "-M", "--move"},
				Subcommands: map[string]*Rule{
					"-l":             {Default: Allow},
					"--list":         {Default: Allow},
					"-a":             {Default: Allow},
					"-r":             {Default: Allow},
					"-v":             {Default: Allow},
					"--verbose":      {Default: Allow},
					"--show-current": {Default: Allow},
					"--contains":     {Default: Allow},
					"--no-contains":  {Default: Allow},
					"--merged":       {Default: Allow},
					"--no-merged":    {Default: Allow},
				},
			},
			"tag": {
				PromptFlags: []string{"-d", "--delete"},
				Subcommands: map[string]*Rule{
					"-l":       {Default: Allow},
					"--list":   {Default: Allow},
					"-n":       {Default: Allow},
					"-v":       {Default: Allow},
					"--verify": {Default: Allow},
				},
			},
			"stash": {
				Subcommands: map[string]*Rule{
					"list": {Default: Allow},
					"show": {Default: Allow},
				},
			},
			"submodule": {
				Subcommands: map[string]*Rule{
					"status":  {Default: Allow},
					"summary": {Default: Allow},
				},
			},
			"worktree": {
				Subcommands: map[string]*Rule{
					"list": {Default: Allow},
				},
			},
			"config": {
				PromptFlags: []string{"--global", "--system", "--file"},
				Subcommands: map[string]*Rule{
					"list":   {Default: Allow},
					"get":    {Default: Allow},
					"--list": {Default: Allow},
					"--get":  {Default: Allow},
				},
			},
		},
	},

	"go": {
		Subcommands: map[string]*Rule{
			"version": {Default: Allow},
			"env":     {Default: Allow},
			"list":    {Default: Allow},
			"doc":     {Default: Allow},
			"vet":     {Default: Allow},
			"test": {
				Default:     Allow,
				PromptFlags: []string{"-c", "--coverprofile"},
			},
			"build": {
				Subcommands: map[string]*Rule{
					"./...": {Default: Allow},
				},
			},
			"mod": {
				Subcommands: map[string]*Rule{
					"download": {Default: Allow},
					"verify":   {Default: Allow},
					"graph":    {Default: Allow},
					"why":      {Default: Allow},
				},
			},
			"tool": {
				Subcommands: map[string]*Rule{
					"vet":   {Default: Allow},
					"cover": {Default: Allow},
					"pprof": {Default: Allow},
					"trace": {Default: Allow},
				},
			},
		},
	},

	"gofmt": {
		Default: Allow,
		AllowFlags: []string{
			"-l", "--list",
			"-d", "--diff",
			"-s", "--simplify",
			"-e",
			"-r", "--rewrite",
			"-cpuprofile",
			"-memprofile",
			"-trace",
		},
	},
	"gofumpt": {
		Default: Allow,
		AllowFlags: []string{
			"-l",
			"-d", "--diff",
			"-e",
			"-cpuprofile",
			"-memprofile",
			"-trace",
		},
	},
	"goimports": {
		Default: Allow,
		AllowFlags: []string{
			"-l",
			"-d",
			"-e",
		},
	},

	"cargo": {
		Subcommands: map[string]*Rule{
			"version":        {Default: Allow},
			"--version":      {Default: Allow},
			"-V":             {Default: Allow},
			"check":          {Default: Allow},
			"test":           {Default: Allow},
			"doc":            {Default: Allow},
			"tree":           {Default: Allow},
			"locate-project": {Default: Allow},
			"metadata":       {Default: Allow},
			"search":         {Default: Allow},
			"info":           {Default: Allow},
		},
	},

	"gh": {
		Subcommands: map[string]*Rule{
			"version": {Default: Allow},
			"pr": {
				Subcommands: map[string]*Rule{
					"list":   {Default: Allow},
					"view":   {Default: Allow},
					"status": {Default: Allow},
					"checks": {Default: Allow},
					"diff":   {Default: Allow},
				},
			},
			"issue": {
				Subcommands: map[string]*Rule{
					"list":   {Default: Allow},
					"view":   {Default: Allow},
					"status": {Default: Allow},
					"search": {Default: Allow},
				},
			},
			"repo": {
				Subcommands: map[string]*Rule{
					"list":   {Default: Allow},
					"view":   {Default: Allow},
					"clone":  {Default: Allow},
					"fork":   {Default: Allow},
					"search": {Default: Allow},
				},
			},
			"release": {
				Subcommands: map[string]*Rule{
					"list":     {Default: Allow},
					"view":     {Default: Allow},
					"download": {Default: Allow},
				},
			},
			"workflow": {
				Subcommands: map[string]*Rule{
					"list": {Default: Allow},
					"view": {Default: Allow},
				},
			},
			"run": {
				Subcommands: map[string]*Rule{
					"list":     {Default: Allow},
					"view":     {Default: Allow},
					"watch":    {Default: Allow},
					"download": {Default: Allow},
				},
			},
			"auth": {
				Subcommands: map[string]*Rule{
					"status": {Default: Allow},
				},
			},
			"gpg-key": {
				Subcommands: map[string]*Rule{
					"list": {Default: Allow},
				},
			},
			"ssh-key": {
				Subcommands: map[string]*Rule{
					"list": {Default: Allow},
				},
			},
		},
	},

	"kubectl": {
		Subcommands: map[string]*Rule{
			"get":           {Default: Allow},
			"describe":      {Default: Allow},
			"explain":       {Default: Allow},
			"logs":          {Default: Allow},
			"top":           {Default: Allow},
			"api-resources": {Default: Allow},
			"api-versions":  {Default: Allow},
			"version":       {Default: Allow},
			"cluster-info":  {Default: Allow},
			"auth": {
				Subcommands: map[string]*Rule{
					"can-i":     {Default: Allow},
					"reconcile": {Default: Allow},
					"whoami":    {Default: Allow},
				},
			},
			"rollout": {
				Default: Prompt,
				Subcommands: map[string]*Rule{
					"status":  {Default: Allow},
					"history": {Default: Allow},
				},
			},
		},
	},

	"docker": {
		Subcommands: map[string]*Rule{
			"version": {Default: Allow},
			"info":    {Default: Allow},
			"images":  {Default: Allow},
			"ps":      {Default: Allow},
			"inspect": {Default: Allow},
			"logs":    {Default: Allow},
			"stats":   {Default: Allow},
			"top":     {Default: Allow},
			"port":    {Default: Allow},
			"history": {Default: Allow},
			"search":  {Default: Allow},
			"volume": {
				Subcommands: map[string]*Rule{
					"ls":      {Default: Allow},
					"inspect": {Default: Allow},
				},
			},
			"network": {
				Subcommands: map[string]*Rule{
					"ls":      {Default: Allow},
					"inspect": {Default: Allow},
				},
			},
			"compose": {
				Default: Prompt,
				Subcommands: map[string]*Rule{
					"ps":   {Default: Allow},
					"logs": {Default: Allow},
				},
			},
		},
	},

	"podman": {
		Subcommands: map[string]*Rule{
			"version": {Default: Allow},
			"info":    {Default: Allow},
			"images":  {Default: Allow},
			"ps":      {Default: Allow},
			"inspect": {Default: Allow},
			"logs":    {Default: Allow},
			"stats":   {Default: Allow},
			"top":     {Default: Allow},
			"port":    {Default: Allow},
			"history": {Default: Allow},
			"search":  {Default: Allow},
			"volume": {
				Subcommands: map[string]*Rule{
					"ls":      {Default: Allow},
					"inspect": {Default: Allow},
				},
			},
			"network": {
				Subcommands: map[string]*Rule{
					"ls":      {Default: Allow},
					"inspect": {Default: Allow},
				},
			},
		},
	},

	// ── JavaScript / TypeScript package managers ────────────
	"bun": {
		Subcommands: map[string]*Rule{
			"--version": {Default: Allow},
			"test":      {Default: Allow},
			"check":     {Default: Allow},
		},
	},

	"pnpm": {
		Subcommands: map[string]*Rule{
			"--version": {Default: Allow},
			"list":      {Default: Allow},
			"ls":        {Default: Allow},
			"why":       {Default: Allow},
			"outdated":  {Default: Allow},
			"audit":     {Default: Allow},
			"info":      {Default: Allow},
		},
	},

	"yarn": {
		Subcommands: map[string]*Rule{
			"--version": {Default: Allow},
			"info":      {Default: Allow},
			"list":      {Default: Allow},
			"why":       {Default: Allow},
			"outdated":  {Default: Allow},
			"dir":       {Default: Allow},
			"bin":       {Default: Allow},
			"npm": {
				Subcommands: map[string]*Rule{
					"info": {Default: Allow},
				},
			},
			"constraints": {
				Subcommands: map[string]*Rule{
					"query": {Default: Allow},
				},
			},
		},
	},

	// ── Python package managers ─────────────────────────────
	"pip": {
		Subcommands: map[string]*Rule{
			"--version": {Default: Allow},
			"list":      {Default: Allow},
			"show":      {Default: Allow},
			"search":    {Default: Allow},
			"check":     {Default: Allow},
			"freeze":    {Default: Allow},
			"deptree":   {Default: Allow},
			"index": {
				Subcommands: map[string]*Rule{
					"versions": {Default: Allow},
				},
			},
			"hash": {Default: Allow},
		},
	},

	"pip3": {
		Subcommands: map[string]*Rule{
			"--version": {Default: Allow},
			"list":      {Default: Allow},
			"show":      {Default: Allow},
			"search":    {Default: Allow},
			"check":     {Default: Allow},
			"freeze":    {Default: Allow},
			"deptree":   {Default: Allow},
			"index": {
				Subcommands: map[string]*Rule{
					"versions": {Default: Allow},
				},
			},
			"hash": {Default: Allow},
		},
	},

	"pipx": {
		Subcommands: map[string]*Rule{
			"--version": {Default: Allow},
			"list":      {Default: Allow},
		},
	},

	"uv": {
		Subcommands: map[string]*Rule{
			"--version": {Default: Allow},
			"pip": {
				Subcommands: map[string]*Rule{
					"list":    {Default: Allow},
					"show":    {Default: Allow},
					"check":   {Default: Allow},
					"freeze":  {Default: Allow},
					"deptree": {Default: Allow},
				},
			},
			"tool": {
				Subcommands: map[string]*Rule{
					"list": {Default: Allow},
				},
			},
			"python": {
				Subcommands: map[string]*Rule{
					"list": {Default: Allow},
					"find": {Default: Allow},
				},
			},
		},
	},

	"poetry": {
		Subcommands: map[string]*Rule{
			"--version": {Default: Allow},
			"show":      {Default: Allow},
			"search":    {Default: Allow},
			"check":     {Default: Allow},
			"env": {
				Subcommands: map[string]*Rule{
					"list": {Default: Allow},
					"info": {Default: Allow},
				},
			},
		},
	},

	// ── Nix ─────────────────────────────────────────────────
	"nix": {
		Subcommands: map[string]*Rule{
			"--version": {Default: Allow},
			"eval":      {Default: Allow},
			"search":    {Default: Allow},
			"hash":      {Default: Allow},
			"path-info": {Default: Allow},
			"flake": {
				Subcommands: map[string]*Rule{
					"metadata": {Default: Allow},
					"check":    {Default: Allow},
					"show":     {Default: Allow},
					"list":     {Default: Allow},
				},
			},
			"store": {
				Subcommands: map[string]*Rule{
					"ls":            {Default: Allow},
					"cat":           {Default: Allow},
					"diff-closures": {Default: Allow},
					"ping":          {Default: Allow},
					"verify":        {Default: Allow},
				},
			},
			"why-depends": {Default: Allow},
		},
	},

	// ── Infrastructure as code ──────────────────────────────
	"terraform": {
		Subcommands: map[string]*Rule{
			"version":   {Default: Allow},
			"validate":  {Default: Allow},
			"plan":      {Default: Allow},
			"show":      {Default: Allow},
			"output":    {Default: Allow},
			"graph":     {Default: Allow},
			"providers": {Default: Allow},
			"workspace": {
				Subcommands: map[string]*Rule{
					"list": {Default: Allow},
					"show": {Default: Allow},
				},
			},
			"state": {
				Subcommands: map[string]*Rule{
					"list": {Default: Allow},
					"show": {Default: Allow},
				},
			},
		},
	},

	"tofu": {
		Subcommands: map[string]*Rule{
			"version":   {Default: Allow},
			"validate":  {Default: Allow},
			"plan":      {Default: Allow},
			"show":      {Default: Allow},
			"output":    {Default: Allow},
			"graph":     {Default: Allow},
			"providers": {Default: Allow},
			"workspace": {
				Subcommands: map[string]*Rule{
					"list": {Default: Allow},
					"show": {Default: Allow},
				},
			},
			"state": {
				Subcommands: map[string]*Rule{
					"list": {Default: Allow},
					"show": {Default: Allow},
				},
			},
		},
	},

	"helm": {
		Subcommands: map[string]*Rule{
			"version": {Default: Allow},
			"list":    {Default: Allow},
			"status":  {Default: Allow},
			"get":     {Default: Allow},
			"search":  {Default: Allow},
			"show":    {Default: Allow},
			"history": {Default: Allow},
			"env":     {Default: Allow},
			"repo": {
				Subcommands: map[string]*Rule{
					"list":   {Default: Allow},
					"search": {Default: Allow},
				},
			},
			"template": {Default: Allow},
		},
	},
}
