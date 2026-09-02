#!/usr/bin/env bash
# kb.sh — query and extend the project knowledge base (docs/knowledge/truf.jsonl).
# One JSON object per line: id, type, topic, tags[], summary, detail, refs[], date.
#
#   kb.sh search <term> [term...]   AND-match (case-insensitive) over the whole entry; one line per hit
#   kb.sh show <id> [id...]         full entry (summary, detail, refs)
#   kb.sh type <type>               entries of a type      (concept decision convention spec howto perf issue gotcha history reference)
#   kb.sh topic <topic>             entries of a topic     (domain storage ui money ci testing dev git docs cli performance code-style)
#   kb.sh tag <tag>                 entries carrying a tag
#   kb.sh list                      every entry, one line
#   kb.sh types | topics | tags     counts
#   kb.sh refs <path-fragment>      entries that reference a file
#   kb.sh add --id X --type T --topic P --summary S [--detail D] [--tags a,b] [--refs p1,p2]
#   kb.sh validate                  JSON well-formed, required fields, unique ids
#   kb.sh check                     validate + every ref exists in the tree (CI gate)
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
KB="${KB_FILE:-$ROOT/docs/knowledge/truf.jsonl}"

jq_bin() {
  if command -v jq >/dev/null 2>&1; then echo jq; return; fi
  for c in "$LOCALAPPDATA/Microsoft/WinGet/Links/jq.exe" "/c/Users/$USER/AppData/Local/Microsoft/WinGet/Links/jq.exe"; do
    [ -x "$c" ] && { echo "$c"; return; }
  done
  echo "kb.sh: jq not found (winget install jqlang.jq)" >&2; exit 127
}
JQ="$(jq_bin)"

# one line per entry: id  [type/topic]  summary
line_fmt='"\(.id)\t[\(.type)/\(.topic)]\t\(.summary)"'

full_fmt='
"── \(.id)  [\(.type)/\(.topic)]  \(.date)\n" +
"   \(.summary)\n" +
(if (.detail // "") != "" then "   · \(.detail)\n" else "" end) +
(if (.tags | length) > 0 then "   tags: \(.tags | join(", "))\n" else "" end) +
(if (.refs | length) > 0 then "   refs: \(.refs | join(", "))\n" else "" end)'

cmd="${1:-help}"; shift || true

case "$cmd" in
  search|s)
    [ $# -ge 1 ] || { echo "usage: kb.sh search <term> [term...]" >&2; exit 2; }
    # each term must appear somewhere in the JSON line (any field), case-insensitive
    filter='.'
    for t in "$@"; do
      esc=$(printf '%s' "$t" | sed 's/[][\\.^$*+?(){}|/]/\\&/g')
      filter="$filter | select(tostring | test(\"$esc\"; \"i\"))"
    done
    "$JQ" -r "$filter | $line_fmt" "$KB"
    ;;
  show)
    [ $# -ge 1 ] || { echo "usage: kb.sh show <id> [id...]" >&2; exit 2; }
    for id in "$@"; do
      "$JQ" -r --arg id "$id" "select(.id == \$id) | $full_fmt" "$KB"
    done
    ;;
  type)   "$JQ" -r --arg v "$1" "select(.type == \$v) | $line_fmt" "$KB" ;;
  topic)  "$JQ" -r --arg v "$1" "select(.topic == \$v) | $line_fmt" "$KB" ;;
  tag)    "$JQ" -r --arg v "$1" "select(.tags | index(\$v)) | $line_fmt" "$KB" ;;
  refs)   "$JQ" -r --arg v "$1" "select(.refs | map(contains(\$v)) | any) | $line_fmt" "$KB" ;;
  list|ls) "$JQ" -r "$line_fmt" "$KB" ;;
  types)  "$JQ" -r '.type'  "$KB" | sort | uniq -c | sort -rn ;;
  topics) "$JQ" -r '.topic' "$KB" | sort | uniq -c | sort -rn ;;
  tags)   "$JQ" -r '.tags[]' "$KB" | sort | uniq -c | sort -rn ;;
  add)
    id=""; type=""; topic=""; summary=""; detail=""; tags=""; refs=""; date="$(date +%F)"
    while [ $# -gt 0 ]; do
      case "$1" in
        --id) id="$2"; shift 2 ;;
        --type) type="$2"; shift 2 ;;
        --topic) topic="$2"; shift 2 ;;
        --summary) summary="$2"; shift 2 ;;
        --detail) detail="$2"; shift 2 ;;
        --tags) tags="$2"; shift 2 ;;
        --refs) refs="$2"; shift 2 ;;
        --date) date="$2"; shift 2 ;;
        *) echo "kb.sh add: unknown flag $1" >&2; exit 2 ;;
      esac
    done
    [ -n "$id" ] && [ -n "$type" ] && [ -n "$topic" ] && [ -n "$summary" ] || {
      echo "usage: kb.sh add --id X --type T --topic P --summary S [--detail D] [--tags a,b] [--refs p1,p2]" >&2; exit 2; }
    if "$JQ" -e --arg id "$id" 'select(.id == $id)' "$KB" >/dev/null 2>&1; then
      echo "kb.sh add: id '$id' already exists (use a new id or edit the line)" >&2; exit 1
    fi
    "$JQ" -cn --arg id "$id" --arg type "$type" --arg topic "$topic" --arg summary "$summary" \
      --arg detail "$detail" --arg tags "$tags" --arg refs "$refs" --arg date "$date" '
      def split_list: if . == "" then [] else split(",") | map(gsub("^\\s+|\\s+$"; "")) end;
      {id: $id, type: $type, topic: $topic, tags: ($tags | split_list),
       summary: $summary, detail: $detail, refs: ($refs | split_list), date: $date}' >> "$KB"
    echo "added $id"
    ;;
  validate)
    n=0; bad=0
    while IFS= read -r line || [ -n "$line" ]; do
      n=$((n+1))
      if ! printf '%s\n' "$line" | "$JQ" -e '
          type == "object" and (.id|type)=="string" and (.type|type)=="string" and (.topic|type)=="string"
          and (.summary|type)=="string" and (.tags|type)=="array" and (.refs|type)=="array" and (.date|type)=="string"
        ' >/dev/null 2>&1; then
        echo "line $n: invalid JSON or missing field" >&2; bad=$((bad+1))
      fi
    done < "$KB"
    dups=$("$JQ" -r '.id' "$KB" | sort | uniq -d)
    if [ -n "$dups" ]; then echo "duplicate ids:" >&2; echo "$dups" >&2; bad=$((bad+1)); fi
    if [ "$bad" -eq 0 ]; then echo "ok: $n entries, ids unique"; else exit 1; fi
    ;;
  check)
    # CI gate: validate + every ref exists in the tree (or is gitignored, e.g. local-only reports/specs)
    "$0" validate
    bad=0
    while IFS=$'\t' read -r id ref; do
      ref="${ref%$'\r'}"
      [ -e "$ROOT/$ref" ] && continue
      if git -C "$ROOT" check-ignore -q "$ref" 2>/dev/null; then continue; fi
      echo "$id: ref '$ref' does not exist" >&2; bad=$((bad+1))
    done < <("$JQ" -r '.id as $id | .refs[] | [$id, .] | @tsv' "$KB")
    if [ "$bad" -eq 0 ]; then echo "ok: all refs resolve"; else exit 1; fi
    ;;
  help|-h|--help|*)
    sed -n '2,16p' "$0" | sed 's/^# \{0,1\}//'
    ;;
esac
