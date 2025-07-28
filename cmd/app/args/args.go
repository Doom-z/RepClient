package args

type Args struct {
	Trial    bool   `arg:"--trial" help:"trial mode" default:"false"`
	Ipv4     string `arg:"-i,--ipv4" help:"ipv4 address to query"`
	Ipv6     string `arg:"--ipv6" help:"ipv6 address to query"`
	Ns       string `arg:"-s,--ns" help:"ns to query"`
	Cname    string `arg:"-n,--cname" help:"cname to query"`
	Txt      string `arg:"-t,--txt" help:"txt to query"`
	Mx       string `arg:"-x,--mx" help:"mx to query"`
	ListFile string `arg:"-l,--list-file" help:"Path to file containing list of DNS entries (ipv4, ipv6, ns, cname, txt, mx); type will be auto-detected"`
	ModeFull bool   `arg:"-f,--full" help:"full mode, fetch all column such as ASN, ASN Name, City, Country, etc" default:"false"`

	MaxTotalOutputIp int    `arg:"-m,--max" help:"max total output per ip" default:"100"`
	PageSize         int    `arg:"-p,--page-size" help:"page size" default:"100"`
	Output           bool   `arg:"-o,--output" help:"output to file" default:"false"`
	Threads          int    `arg:"-t,--threads" help:"number of threads" default:"1"`
	Verbose          bool   `arg:"-v,--verbose" help:"verbose output" default:"false"`
	Config           string `arg:"-c,--config" help:"config file" default:"config.toml"`
}
