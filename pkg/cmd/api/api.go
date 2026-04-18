package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/itchyny/gojq"
	"github.com/spf13/cobra"

	"github.com/CircleCI-Public/circleci-cli/pkg/cmdutil"
	cierrors "github.com/CircleCI-Public/circleci-cli/pkg/errors"
)

// NewCmdAPI returns `circleci api <endpoint>`.
func NewCmdAPI(f *cmdutil.Factory) *cobra.Command {
	var method string
	var fields []string
	var headers []string
	var paginate bool
	var jqExpr string

	cmd := &cobra.Command{
		Use:   "api <endpoint>",
		Short: "Make authenticated requests to the CircleCI API",
		Long: heredoc.Doc(`
			Send an authenticated HTTP request to any CircleCI API endpoint and
			print the response as JSON.

			The endpoint must be a path starting with / (e.g. /me or /project/gh/myorg/myrepo).
			The base URL and authentication header are added automatically.

			Use --field to send request body fields (JSON encoded). Use --header to
			add custom HTTP headers. Use --paginate to follow next-page tokens and
			collect all results. Use --jq to filter the final JSON output.
		`),
		Example: heredoc.Doc(`
			# Get the authenticated user:
			$ circleci api /me

			# List pipelines for a project:
			$ circleci api /project/gh/myorg/myrepo/pipeline --method GET

			# Create a context:
			$ circleci api /context --method POST \
			    --field name=my-context \
			    --field owner.id=$ORG_ID \
			    --field owner.type=organization

			# Paginate through all results:
			$ circleci api /project/gh/myorg/myrepo/pipeline --paginate

			# Filter with jq:
			$ circleci api /project/gh/myorg/myrepo/pipeline --jq '.items[].id'
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint := args[0]
			if !strings.HasPrefix(endpoint, "/") {
				endpoint = "/" + endpoint
			}

			// Build body from --field flags.
			body := map[string]interface{}{}
			for _, f := range fields {
				parts := strings.SplitN(f, "=", 2)
				if len(parts) != 2 {
					return cierrors.New("INVALID_ARG",
						fmt.Sprintf("Invalid --field %q", f),
						"Use the form key=value.", cierrors.ExitBadArguments)
				}
				setNestedField(body, parts[0], parts[1])
			}

			// Build extra headers.
			extraHeaders := map[string]string{}
			for _, h := range headers {
				parts := strings.SplitN(h, ":", 2)
				if len(parts) != 2 {
					return cierrors.New("INVALID_ARG",
						fmt.Sprintf("Invalid --header %q", h),
						"Use the form 'Header-Name: value'.", cierrors.ExitBadArguments)
				}
				extraHeaders[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			if method == "" {
				if len(body) > 0 {
					method = http.MethodPost
				} else {
					method = http.MethodGet
				}
			}
			method = strings.ToUpper(method)

			var results []interface{}

			pageToken := ""
			for {
				path := endpoint
				if paginate && pageToken != "" {
					sep := "?"
					if strings.Contains(path, "?") {
						sep = "&"
					}
					path = path + sep + "page-token=" + pageToken
				}

				stop := f.IOStreams.StartSpinner("Calling API...")
				respData, err := client.RawRequest(method, path, body, extraHeaders)
				stop()
				if err != nil {
					return err
				}

				var parsed interface{}
				if err := json.Unmarshal(respData, &parsed); err != nil {
					// Not JSON — print raw.
					fmt.Fprintln(f.IOStreams.Out, string(respData))
					return nil
				}

				if !paginate {
					return printJSON(f, parsed, jqExpr)
				}

				// Pagination: collect items from paged responses.
				if m, ok := parsed.(map[string]interface{}); ok {
					if items, ok := m["items"].([]interface{}); ok {
						results = append(results, items...)
					} else {
						results = append(results, parsed)
					}
					nextToken, _ := m["next_page_token"].(string)
					if nextToken == "" {
						break
					}
					pageToken = nextToken
				} else {
					results = append(results, parsed)
					break
				}
			}

			if paginate {
				return printJSON(f, results, jqExpr)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&method, "method", "X", "", "HTTP method (default: GET, or POST when --field is used)")
	cmd.Flags().StringArrayVarP(&fields, "field", "F", nil, "Request body field in key=value form (repeatable)")
	cmd.Flags().StringArrayVarP(&headers, "header", "H", nil, "Additional HTTP header in 'Name: value' form (repeatable)")
	cmd.Flags().BoolVar(&paginate, "paginate", false, "Follow next_page_token and collect all pages")
	cmd.Flags().StringVar(&jqExpr, "jq", "", "Filter output with a jq expression")
	return cmd
}

// setNestedField sets body["a"]["b"] for dotted keys like "owner.id".
func setNestedField(body map[string]interface{}, key, value string) {
	parts := strings.SplitN(key, ".", 2)
	if len(parts) == 1 {
		body[key] = value
		return
	}
	sub, ok := body[parts[0]].(map[string]interface{})
	if !ok {
		sub = map[string]interface{}{}
		body[parts[0]] = sub
	}
	setNestedField(sub, parts[1], value)
}

func printJSON(f *cmdutil.Factory, data interface{}, jqExpr string) error {
	if jqExpr != "" {
		q, err := gojq.Parse(jqExpr)
		if err != nil {
			return cierrors.New("INVALID_ARG", "Invalid --jq expression",
				err.Error(), cierrors.ExitBadArguments)
		}
		iter := q.Run(data)
		for {
			v, ok := iter.Next()
			if !ok {
				break
			}
			if err, ok := v.(error); ok {
				return cierrors.New("JQ_ERROR", "jq error", err.Error(), cierrors.ExitGeneralError)
			}
			out, _ := json.MarshalIndent(v, "", "  ")
			fmt.Fprintln(f.IOStreams.Out, string(out))
		}
		return nil
	}

	out, _ := json.MarshalIndent(data, "", "  ")
	fmt.Fprintln(f.IOStreams.Out, string(out))
	return nil
}

