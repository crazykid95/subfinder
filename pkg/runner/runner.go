package runner

import (
	"bufio"
	"io"
	"os"
	"path"

	"github.com/projectdiscovery/subfinder/pkg/passive"
	"github.com/projectdiscovery/subfinder/pkg/resolve"
)

// Runner is an instance of the subdomain enumeration
// client used to orchestrate the whole process.
type Runner struct {
	options        *Options
	passiveAgent   *passive.Agent
	resolverClient *resolve.Resolver
}

// NewRunner creates a new runner struct instance by parsing
// the configuration options, configuring sources, reading lists
// and setting up loggers, etc.
func NewRunner(options *Options) (*Runner, error) {
	runner := &Runner{options: options}

	// Initialize the passive subdomain enumeration engine
	runner.initializePassiveEngine()

	// Initialize the active subdomain enumeration engine
	err := runner.initializeActiveEngine()
	if err != nil {
		return nil, err
	}

	return runner, nil
}

// RunEnumeration runs the subdomain enumeration flow on the targets specified
func (r *Runner) RunEnumeration() error {
	// Check if only a single domain is sent as input. Process the domain now.
	if r.options.Domain != "" {
		return r.EnumerateSingleDomain(r.options.Domain, r.options.Output, false)
	}

	// If we have multiple domains as input,
	if r.options.DomainsFile != "" {
		f, err := os.Open(r.options.DomainsFile)
		if err != nil {
			return err
		}
		err = r.EnumerateMultipleDomains(f)
		f.Close()
		return err
	}

	// If we have STDIN input, treat it as multiple domains
	if r.options.Stdin {
		return r.EnumerateMultipleDomains(os.Stdin)
	}
	return nil
}

// EnumerateMultipleDomains enumerates subdomains for multiple domains
// We keep enumerating subdomains for a given domain until we reach an error
func (r *Runner) EnumerateMultipleDomains(reader io.Reader) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		domain := scanner.Text()
		if domain == "" {
			continue
		}

		var err error
		// If the user has specifed an output file, use that output file instead
		// of creating a new output file for each domain. Else create a new file
		// for each domain in the directory.
		if r.options.Output != "" {
			err = r.EnumerateSingleDomain(domain, r.options.Output, true)
		} else {
			outputFile := path.Join(r.options.OutputDirectory, domain)
			err = r.EnumerateSingleDomain(domain, outputFile, false)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
