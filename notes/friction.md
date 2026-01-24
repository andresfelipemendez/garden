# Eliminate friction
Repetitive work irritates me quite easily so I try to optimize and automate as much as I can so while working on my [[eink-dashboard]] project I noticed that I avoid setting up my Unix environments with my [dotfiles](https://github.com/andresfelipemendez/dotfiles) since it requires a few extra manual steps so I decided to add a [`install.sh`](https://github.com/andresfelipemendez/dotfiles/blob/main/install.sh) to my repo inspired by [on-my-zsh](https://ohmyz.sh/#install).

# automate the iteration
to bootstrap writing the script I decided to start with vim in WSL Ubuntu so the first steps is to add temporary remaps while I get the script to install Neovim and LazyVim I've documented the questions I had to search while doing this:

**Q:** how to open a terminal like VS Code?
**A:** `split | terminal`

**Q:** How to exit terminal mode?
**A:** `Ctrl+w N` that's capital N so `Shift+N`  
  
**Q:** The terminal opened above my buffer, how do I switch them so it's at the bottom?
**A:** `Ctrl+w r`

# Temporary shortcuts
Since I'll be switching constantly between these buffers I added a shortcut to save and switch
```
:tnoremap <S-TAb> <C-w>N<C-w>k
:nnoremap <silent> <S-TAb> :w<CR><C-w>j:call feedkeys('i')<CR>
:inoremap <silent> <S-TAb> <Esc>:w<CR><C-w>j:call feedkeys('i')<CR>
```

# `install.sh` Wish-list
When I'm done it should detect the OS to support Debian/Ubuntu for my home-lab and OSX for my laptop. I want a consistent behavior as it is possible so muscle memory transfers well across these envs so similar tools similar shortcuts