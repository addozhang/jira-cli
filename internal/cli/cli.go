package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/addozhang/jira-cli/internal/app"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type Runtime struct {
	Out             io.Writer
	Err             io.Writer
	In              io.Reader
	CredentialsPath string
	Output          string
	Timeout         time.Duration
	Insecure        bool
	Debug           bool
}

func Execute() int {
	rt := &Runtime{Out: os.Stdout, Err: os.Stderr, In: os.Stdin}
	cmd := NewRootCommand(rt)
	if err := cmd.Execute(); err != nil {
		var appErr *app.Error
		if errors.As(err, &appErr) {
			_, _ = fmt.Fprintln(rt.Err, appErr.Message)
			if appErr.Next != "" {
				_, _ = fmt.Fprintln(rt.Err, "Next: "+appErr.Next)
			}
			return app.ErrorExitCode
		}
		_, _ = fmt.Fprintln(rt.Err, err)
		return app.ErrorExitCode
	}
	return 0
}

func NewRootCommand(rt *Runtime) *cobra.Command {
	if rt.Out == nil {
		rt.Out = io.Discard
	}
	if rt.Err == nil {
		rt.Err = io.Discard
	}
	if rt.In == nil {
		rt.In = strings.NewReader("")
	}
	root := &cobra.Command{Use: "jr", Short: "URL-native Jira Server/Data Center CLI", SilenceUsage: true, SilenceErrors: true}
	root.SetOut(rt.Out)
	root.SetErr(rt.Err)
	root.PersistentFlags().StringVarP(&rt.Output, "output", "o", "yaml", "output format: yaml, json, raw")
	root.PersistentFlags().DurationVar(&rt.Timeout, "timeout", 30*time.Second, "per-request timeout")
	root.PersistentFlags().BoolVar(&rt.Insecure, "insecure", false, "disable TLS certificate verification")
	root.PersistentFlags().BoolVar(&rt.Debug, "debug", false, "print redacted HTTP exchange to stderr")
	root.AddCommand(versionCommand(rt), authCommand(rt), issueCommand(rt), searchCommand(rt))
	return root
}

func versionCommand(rt *Runtime) *cobra.Command {
	return &cobra.Command{Use: "version", Short: "Print version information", RunE: func(cmd *cobra.Command, args []string) error {
		return app.Render(rt.Out, rt.Output, map[string]string{"schemaVersion": app.SchemaVersion, "version": version, "commit": commit, "date": date})
	}}
}

func authCommand(rt *Runtime) *cobra.Command {
	cmd := &cobra.Command{Use: "auth", Short: "Manage Jira credentials"}
	var alias string
	add := &cobra.Command{Use: "add <url>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		store, err := loadStore(rt)
		if err != nil {
			return err
		}
		_, _ = fmt.Fprint(rt.Err, "Token: ")
		token, err := readSecret(rt.In)
		_, _ = fmt.Fprintln(rt.Err)
		if err != nil {
			return app.WrapError("Could not read token", "Paste a Jira Personal Access Token and try again.", err)
		}
		key, err := store.Add(args[0], strings.TrimSpace(token), alias)
		if err != nil {
			return err
		}
		return app.Render(rt.Out, rt.Output, map[string]any{"schemaVersion": app.SchemaVersion, "instance": key, "alias": alias})
	}}
	add.Flags().StringVar(&alias, "alias", "", "short instance alias")

	list := &cobra.Command{Use: "list", RunE: func(cmd *cobra.Command, args []string) error {
		store, err := loadStore(rt)
		if err != nil {
			return err
		}
		return app.Render(rt.Out, rt.Output, map[string]any{"schemaVersion": app.SchemaVersion, "instances": keys(store.Data.Instances), "aliases": store.Data.Aliases})
	}}
	remove := &cobra.Command{Use: "remove <url-or-alias>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		store, err := loadStore(rt)
		if err != nil {
			return err
		}
		removed, err := store.Remove(args[0])
		if err != nil {
			return err
		}
		return app.Render(rt.Out, rt.Output, map[string]any{"schemaVersion": app.SchemaVersion, "removed": removed})
	}}
	whoami := &cobra.Command{Use: "whoami <url-or-alias>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		client, _, err := clientForInstance(rt, args[0])
		if err != nil {
			return err
		}
		value, err := client.WhoAmI()
		if err != nil {
			return err
		}
		return app.Render(rt.Out, rt.Output, value)
	}}
	cmd.AddCommand(add, list, remove, whoami)
	return cmd
}

func issueCommand(rt *Runtime) *cobra.Command {
	cmd := &cobra.Command{Use: "issue", Short: "Read and comment on Jira issues"}
	var instance string
	get := &cobra.Command{Use: "get <url-or-key>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		client, target, err := clientForIssue(rt, args[0], instance)
		if err != nil {
			return err
		}
		value, err := client.GetIssue(target.Key, rt.Output == "raw")
		if err != nil {
			return err
		}
		return app.Render(rt.Out, rt.Output, value)
	}}
	get.Flags().StringVarP(&instance, "instance", "i", "", "Jira instance URL or alias for bare issue keys")

	var bodyInput string
	var commentInstance string
	comment := &cobra.Command{Use: "comment <url-or-key>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		body, err := app.ReadBody(bodyInput, rt.In)
		if err != nil {
			return err
		}
		client, target, err := clientForIssue(rt, args[0], commentInstance)
		if err != nil {
			return err
		}
		value, err := client.AddComment(target.Key, body)
		if err != nil {
			return err
		}
		return app.Render(rt.Out, rt.Output, value)
	}}
	comment.Flags().StringVarP(&commentInstance, "instance", "i", "", "Jira instance URL or alias for bare issue keys")
	comment.Flags().StringVar(&bodyInput, "body", "", "comment body: literal text, @path, or -")
	cmd.AddCommand(get, comment)
	return cmd
}

func searchCommand(rt *Runtime) *cobra.Command {
	var instance, jql string
	cmd := &cobra.Command{Use: "search", Short: "Run explicit Jira JQL", RunE: func(cmd *cobra.Command, args []string) error {
		if jql == "" {
			return app.NewError("Missing JQL query", "Pass --jql '<query>'.")
		}
		if instance == "" {
			return app.NewError("Search requires an instance", "Pass --instance URL|alias.")
		}
		client, _, err := clientForInstance(rt, instance)
		if err != nil {
			return err
		}
		value, err := client.Search(jql)
		if err != nil {
			return err
		}
		return app.Render(rt.Out, rt.Output, value)
	}}
	cmd.Flags().StringVar(&jql, "jql", "", "Jira Query Language query")
	cmd.Flags().StringVarP(&instance, "instance", "i", "", "Jira instance URL or alias")
	return cmd
}

func loadStore(rt *Runtime) (*app.CredentialStore, error) {
	path := rt.CredentialsPath
	if path == "" {
		var err error
		path, err = app.DefaultCredentialsPath()
		if err != nil {
			return nil, err
		}
	}
	return app.LoadCredentials(path)
}

func clientForInstance(rt *Runtime, instance string) (app.JiraClient, string, error) {
	store, err := loadStore(rt)
	if err != nil {
		return app.JiraClient{}, "", err
	}
	key, cred, err := store.ResolveInstance(instance)
	if err != nil {
		return app.JiraClient{}, "", err
	}
	httpClient, err := app.NewHTTPClient(app.HTTPOptions{Timeout: rt.Timeout, Insecure: rt.Insecure, Debug: rt.Debug, DebugOut: rt.Err})
	if err != nil {
		return app.JiraClient{}, "", err
	}
	return app.JiraClient{BaseURL: key, Token: cred.Token, HTTP: httpClient}, key, nil
}

func clientForIssue(rt *Runtime, targetArg, instance string) (app.JiraClient, app.IssueTarget, error) {
	store, err := loadStore(rt)
	if err != nil {
		return app.JiraClient{}, app.IssueTarget{}, err
	}
	target, cred, err := app.ResolveIssueTarget(targetArg, instance, store)
	if err != nil {
		return app.JiraClient{}, app.IssueTarget{}, err
	}
	httpClient, err := app.NewHTTPClient(app.HTTPOptions{Timeout: rt.Timeout, Insecure: rt.Insecure, Debug: rt.Debug, DebugOut: rt.Err})
	if err != nil {
		return app.JiraClient{}, app.IssueTarget{}, err
	}
	return app.JiraClient{BaseURL: target.Instance, Token: cred.Token, HTTP: httpClient}, target, nil
}

func readSecret(in io.Reader) (string, error) {
	if f, ok := in.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		data, err := term.ReadPassword(int(f.Fd()))
		return string(data), err
	}
	data, err := io.ReadAll(in)
	return string(data), err
}

func keys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for key := range m {
		out = append(out, key)
	}
	return out
}
