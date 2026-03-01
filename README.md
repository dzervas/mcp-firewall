<p align="center" width="100%">
  <img height="256px" src="./logo.png" />
</p>

# MCP FIREW🔒️LL

This is a small tool that sits between the agent and all tool use requests and is able
to apply regex-based policies per folder, git repo and user.

It currently supports Claude Code and GitHub Copilot CLI through the pretooluse hook.

## Quickstart

Download and install the [release](../../releases/latest) binary somewhere that is accessible from your `$PATH` environment variable.
For more installtion instructions check the [Installation](#installation) section

Add the required snippet to your agent of choice:

<details><summary>Claude Code</summary>

In either `~/.config/settings.json` (global) or `.config/settings.json` (per-project):

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "/usr/local/bin/mcp-firewall claude"
          }
        ]
      }
    ]
  }
}
```

</details>
<details><summary>GitHub Copilot CLI</summary>

In `.github/hooks/mcp-firewall.json` (per-project):

```json
{
  "version": 1,
  "hooks": {
    "preToolUse": [
      {
        "type": "command",
        "command": "/usr/local/bin/mcp-firewall claude"
      }
    ]
  }
}
```

</details>

Then all you have to do is write your first policy. Here's a good starting point:

`~/.config/mcp-firewall/config.jsonnet`

```jsonnet
[
	{
		name: 'Simple commands',
		// Note the space at the end of the patterns!
		// Without it, commands like 'sortmalliciously' would also be allowed!
		allow: [
			'echo ',
			'sort ',
			'uniq ',
			'wc ',
			'ls( -\w+)?$', // Allow ls, ls -lah, etc. but not ls /etc/secrets!
		],
	}
]
```

<br />

> [!TIP]
> While mcp-firewall uses [`jsonnet`](https://jsonnet.org) for all the policy files, it's done only to allow for more
> complex and shared policies. If you're not familiar with the language, treat it as normal JSON
> with the added benefit of supporting comments!

## Installation

There are (currently) 3 ways to download and install mcp-firewall:

- Download the latest compiled binary from the [releases](../../releases/latest)
- Clone and build the project using `go build ./src`
- Use the nix flake

More installation options are coming soon!

## Advanced Usage

For users that want to expand a bit further and utilize jsonnet for shared rulesets across
projects, here are some useful info:

- The `lib` subdirectory of `~/.config/mcp-firewall` or the value of `$MCP_FIREWALL_CONFIG_DIR` can be used for libsonnet files
- The `vendor` subdirectory under the afformentioned directories can be used for vendored libraries (using jsonnet-bundler for example)
- The used jsonnet implementation is go-jsonnet, you can see which version in [go.mod](./src/go.mod)
