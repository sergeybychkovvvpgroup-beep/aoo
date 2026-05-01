bash -c '
set -e

SCRIPT_PATH="$HOME/notes/aoo.sh"

echo "[1/3] install deps..."

if command -v apt >/dev/null; then
#  sudo apt update -y
  sudo apt install -y python3 python3-yaml curl
fi

if ! command -v gum >/dev/null; then
  echo "[2/3] install gum..."
  tmp=$(mktemp -d)
  curl -fsSL "https://github.com/charmbracelet/gum/releases/latest/download/gum_$(uname -s)_$(uname -m).tar.gz" \
    | tar -xz -C "$tmp"
  sudo mv "$tmp"/gum*/gum /usr/local/bin/
fi

echo "[3/3] install script..."

mkdir -p "$HOME/notes"

cat > "$SCRIPT_PATH" <<'"'"'EOF'"'"'
aoo() {
  local db selected id action payload_b64 payload desc

  db="$(mktemp)"

  python3 > "$db" <<'"'"'PY'"'"'
import os, yaml, re, base64

root = os.path.expanduser("~/notes")

def one_line(text, limit=70):
    text = str(text or "").strip()
    text = re.sub(r"\s+", " ", text)
    return text if len(text) <= limit else text[:limit-3] + "..."

idx = 0

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

            idx += 1

            if run:
                payload = base64.b64encode(run.encode()).decode()
                display = f"RUN | {one_line(run, 70)} | ({desc})"
                print(f"{idx}\tRUN\t{payload}\t{desc}\t{display}")

            elif note is not None:
                note_text = str(note).strip()
                if note_text:
                    payload = base64.b64encode(note_text.encode()).decode()
                    display = f"NOTE | {one_line(note_text, 70)} | ({desc})"
                    print(f"{idx}\tNOTE\t{payload}\t{desc}\t{display}")
PY

  if [[ ! -s "$db" ]]; then
    echo "no yaml notes found in ~/notes"
    rm -f "$db"
    return
  fi

  selected="$(
    awk -F "\t" '"'"'{print $1 " │ " $5}'"'"' "$db" |
    gum filter \
      --height 12 \
      --no-fuzzy \
      --prompt "notes> " \
      --placeholder "Search..."
  )"

  [ -z "$selected" ] && {
    rm -f "$db"
    return
  }

  id="$(printf "%s" "$selected" | awk -F " │ " '"'"'{print $1}'"'"')"

  action="$(awk -F "\t" -v id="$id" '"'"'$1 == id {print $2; exit}'"'"' "$db")"
  payload_b64="$(awk -F "\t" -v id="$id" '"'"'$1 == id {print $3; exit}'"'"' "$db")"
  desc="$(awk -F "\t" -v id="$id" '"'"'$1 == id {print $4; exit}'"'"' "$db")"

  payload="$(printf "%s" "$payload_b64" | base64 -d)"

  rm -f "$db"

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
read -r -p "Source now? (y/n): " ans
if [[ "$ans" =~ ^[Yy]$ ]]; then
  source "$SCRIPT_PATH"
  echo "✓ loaded"
fi

echo
read -r -p "Add source to ~/.zshrc? (y/n): " ans
if [[ "$ans" =~ ^[Yy]$ ]]; then
  if ! grep -Fq "source $SCRIPT_PATH" "$HOME/.zshrc" 2>/dev/null; then
    echo "source $SCRIPT_PATH" >> "$HOME/.zshrc"
    echo "✓ added to ~/.zshrc"
  else
    echo "already in ~/.zshrc"
  fi
fi

echo
echo "done"
echo "manual load: source $SCRIPT_PATH"
echo "run: aoo"
'
