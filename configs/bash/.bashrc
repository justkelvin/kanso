# Kanso Minimal ~/.bashrc
# Clean, fast, distraction-free Bash environment

# If not running interactively, don't do anything
[[ $- != *i* ]] && return

# Path additions
export PATH="$HOME/.local/bin:$HOME/go/bin:$PATH"

# History Configuration
HISTCONTROL=ignoreboth:erasedups
HISTSIZE=10000
HISTFILESIZE=20000
shopt -s histappend
shopt -s checkwinsize

# Bash completion
if [ -f /usr/share/bash-completion/bash_completion ]; then
    . /usr/share/bash-completion/bash_completion
elif [ -f /etc/bash_completion ]; then
    . /etc/bash_completion
fi

# Git branch in prompt (pure bash helper)
git_branch() {
    local branch
    branch=$(git symbolic-ref --short HEAD 2>/dev/null || git rev-parse --short HEAD 2>/dev/null)
    if [ -n "$branch" ]; then
        echo -e " \033[38;5;180m($branch)\033[0m"
    fi
}

# Prompt: user@host in blue, directory in green, git branch in orange
# Tokyo Night color accented prompt
PS1='\[\033[38;5;111m\]\u@\h\[\033[00m\]:\[\033[38;5;120m\]\w\[\033[00m\]$(git_branch) \[\033[38;5;111m\]\$\[\033[00m\] '

# Aliases
alias ls='ls --color=auto'
alias ll='ls -lah --color=auto'
alias la='ls -A --color=auto'
alias grep='grep --color=auto'
alias diff='diff --color=auto'
alias ip='ip -color=auto'

# Git shortcuts
alias gs='git status -s'
alias ga='git add'
alias gc='git commit'
alias gp='git push'
alias gl='git log --oneline --graph --decorate'

# Pacman helpers
alias pacs='sudo pacman -S'
alias pacu='sudo pacman -Syu'
alias pacr='sudo pacman -Rns'
alias pacq='pacman -Qe'
