bash -c '
set -e

SCRIPT_PATH="$HOME/notes/aoo.sh"

echo "[1/3] install deps..."

if command -v apt >/dev/null; then
  #sudo apt update -y
  sudo apt install -y python3 python3-yaml curl
fi

if ! command -v gum >/dev/null; then
  echo "[2/3] install gum..."
  tmp=$(mktemp -d)
  curl -fsSL https://github.com/charmbracelet/gum/releases/latest/download/gum_$(uname -s)_$(uname -m).tar.gz \
  | tar -xz -C "$tmp"
  sudo mv "$tmp"/gum*/gum /usr/local/bin/
fi

echo "[3/3] install script..."

mkdir -p "$HOME/notes"

cat <<'\''EOF'\'' > "$SCRIPT_PATH"
aoo() {
  local raw selected meta action payload_b64 payload desc

  raw="$(
python3 <<'\''PY'\''
import os, yaml, re, base64

root = os.path.expanduser("~/notes")

def one_line(text, limit=70):
    text = str(text or "").strip()
    text = re.sub(r"\s+", " ", text)
    return text if len(text) <= limit else text[:limit-3] + "..."

for dirpath, dirnames, filenames in os.walk(root):
    dirnames[:] = [d for d in dirnames if d not in [".git", ".obsidian", "_archive"]]

    for filename in filenames:
        if not filename.endswith((".yml", ".yaml")):
            continue

        path = os.path.join(dirpath, filename)

        try:
            with open(path, "r", encoding="utf-8") as f:
                data = yaml.safe_load(f)
        except Exception:
            continue

        if isinstance(data, dict):
            data = [data]

        if not isinstance(data, list):
            continue

        for item in data:
            if not isinstance(item, dict):
                continue

            desc = str(item.get("desc", "") or "").strip()
            run = str(item.get("run", "") or "").strip()
            note = item.get("note", None)

            if not desc:
                continue

            if run:
                display = f"RUN | {one_line(run, 70)} | ({desc})"
                payload = base64.b64encode(run.encode()).decode()
                print(f"{display}\tRUN\t{payload}\t{desc}")

            elif note is not None:
                note_text = str(note).strip()
                if note_text:
                    display = f"NOTE | {one_line(note_text, 70)} | ({desc})"
                    payload = base64.b64encode(note_text.encode()).decode()
                    print(f"{display}\tNOTE\t{payload}\t{desc}")
PY
  )"

  if [[ -z "$raw" ]]; then
    echo "no yaml notes found in ~/notes"
    return
  fi

  selected="$(
    printf "%s\n" "$raw" |
    cut -f1 |
    gum filter \
      --height 12 \
      --no-fuzzy \
      --prompt "notes> " \
      --placeholder "Search..."
  )"

  [ -z "$selected" ] && return

  meta="$(printf "%s\n" "$raw" | awk -F '\t' -v s="$selected" '\''$1 == s {print $2 "\t" $3 "\t" $4; exit}'\'')"

  action="$(printf "%s" "$meta" | cut -f1)"
  payload_b64="$(printf "%s" "$meta" | cut -f2)"
  desc="$(printf "%s" "$meta" | cut -f3)"

  payload="$(printf "%s" "$payload_b64" | base64 -d)"

  if [[ "$action" == "RUN" ]]; then
    eval "$payload"
  else
    printf "\n- desc: %s\n  note: |\n" "$desc"
    printf "%s\n" "$payload" | sed "s/^/    /"
    printf "\n"
  fi
}
EOF

echo
read -p "Source now? (y/n): " ans
if [[ "$ans" =~ ^[Yy]$ ]]; then
  source "$SCRIPT_PATH"
  echo "✓ loaded"
fi

echo
read -p "Add to ~/.zshrc? (y/n): " ans
if [[ "$ans" =~ ^[Yy]$ ]]; then
  if ! grep -q "source $SCRIPT_PATH" "$HOME/.zshrc"; then
    echo "source $SCRIPT_PATH" >> "$HOME/.zshrc"
    echo "✓ added to .zshrc"
  else
    echo "already in .zshrc"
  fi
fi

echo
echo "done → use: aoo"
'
