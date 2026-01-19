package cli

import (
	"fmt"
	"os"
)

const bashCompletion = `# yams bash completion
_yams() {
    local cur prev words cword
    _init_completion || return

    local commands="status server dump sim principals resources actions accounts policies version completion"

    if [[ ${cword} -eq 1 ]]; then
        COMPREPLY=($(compgen -W "${commands}" -- "${cur}"))
        return
    fi

    local cmd="${words[1]}"
    case "${cmd}" in
        status)
            COMPREPLY=($(compgen -W "-s --server --format" -- "${cur}"))
            ;;
        server)
            COMPREPLY=($(compgen -W "-a --addr -s --source -r --refresh -e --env" -- "${cur}"))
            ;;
        dump)
            COMPREPLY=($(compgen -W "-t --target -o --out -a --aggregator -r --rtype --dry-run" -- "${cur}"))
            if [[ "${prev}" == "-t" || "${prev}" == "--target" ]]; then
                COMPREPLY=($(compgen -W "config org" -- "${cur}"))
            fi
            ;;
        sim)
            COMPREPLY=($(compgen -W "-s --server -p --principal -a --action -r --resource -c --context -o --overlay -x --exact -e --explain -t --trace" -- "${cur}"))
            ;;
        principals|resources|actions|accounts|policies)
            COMPREPLY=($(compgen -W "-s --server -q --query -k --key -f --freeze --format" -- "${cur}"))
            ;;
        completion)
            COMPREPLY=($(compgen -W "bash zsh" -- "${cur}"))
            ;;
    esac
}

complete -F _yams yams
`

const zshCompletion = `#compdef yams

_yams() {
    local -a commands
    commands=(
        'status:Show server status and loaded data sources'
        'server:Start the yams API server'
        'dump:Export AWS organization or config data'
        'sim:Simulate IAM permission checks'
        'principals:List or search IAM principals'
        'resources:List or search AWS resources'
        'actions:List or search IAM actions'
        'accounts:List or search AWS accounts'
        'policies:List or search IAM policies'
        'version:Show version information'
        'completion:Generate shell completion scripts'
    )

    _arguments -C \
        '1:command:->command' \
        '*::arg:->args'

    case "$state" in
        command)
            _describe -t commands 'yams commands' commands
            ;;
        args)
            case "${words[1]}" in
                status)
                    _arguments \
                        '(-s --server)'{-s,--server}'[Server address]:address:' \
                        '--format[Output format]:format:(json table)'
                    ;;
                server)
                    _arguments \
                        '(-a --addr)'{-a,--addr}'[Listen address]:address:' \
                        '*'{-s,--source}'[Data source]:source:_files' \
                        '(-r --refresh)'{-r,--refresh}'[Refresh interval]:seconds:' \
                        '*'{-e,--env}'[Environment variables]:var:'
                    ;;
                dump)
                    _arguments \
                        '(-t --target)'{-t,--target}'[Dump target]:target:(config org)' \
                        '(-o --out)'{-o,--out}'[Output destination]:destination:_files' \
                        '(-a --aggregator)'{-a,--aggregator}'[AWS Config aggregator]:aggregator:' \
                        '*'{-r,--rtype}'[Resource types]:type:' \
                        '(-n --dry-run)'{-n,--dry-run}'[Show what would be done]'
                    ;;
                sim)
                    _arguments \
                        '(-s --server)'{-s,--server}'[Server address]:address:' \
                        '(-p --principal)'{-p,--principal}'[Principal ARN]:arn:' \
                        '(-a --action)'{-a,--action}'[AWS action]:action:' \
                        '(-r --resource)'{-r,--resource}'[Resource ARN]:arn:' \
                        '*'{-c,--context}'[Context key=value]:context:' \
                        '*'{-o,--overlay}'[Overlay file]:file:_files' \
                        '(-x --exact)'{-x,--exact}'[Disable fuzzy matching]' \
                        '(-e --explain)'{-e,--explain}'[Show explanation]' \
                        '(-t --trace)'{-t,--trace}'[Show trace]'
                    ;;
                principals|resources|actions|accounts|policies)
                    _arguments \
                        '(-s --server)'{-s,--server}'[Server address]:address:' \
                        '(-q --query)'{-q,--query}'[Search query]:query:' \
                        '(-k --key)'{-k,--key}'[Primary key]:key:' \
                        '(-f --freeze)'{-f,--freeze}'[Freeze entity]' \
                        '--format[Output format]:format:(json table)'
                    ;;
                completion)
                    _arguments '1:shell:(bash zsh)'
                    ;;
            esac
            ;;
    esac
}

_yams
`

// PrintCompletion outputs shell completion scripts
func PrintCompletion(shell string) {
	switch shell {
	case "bash":
		fmt.Print(bashCompletion)
	case "zsh":
		fmt.Print(zshCompletion)
	default:
		fmt.Fprintf(os.Stderr, "Unknown shell: %s. Supported shells: bash, zsh\n", shell)
		os.Exit(1)
	}
}
