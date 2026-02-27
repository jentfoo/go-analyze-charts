#!/usr/bin/env bash
# eval_range_commits.sh - Run range evaluation matrix across commits and produce a comparison report.
#
# Usage:  ./eval_range_commits.sh
#
# The script creates temporary git worktrees, injects the current eval test file,
# runs TestRangeEvalMatrix, captures results, then cleans up.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# ---- Commits to evaluate ----
# Format: "hash:label"
COMMITS=(
    "b887f7b9daa8bd2da210c1269e392954f6c6f81b:improve_intervals"
    "8f7ba7e9671443a8dcf2841e4db9e0f7468a9841:refactor"
)

SINGLE_CONFIGS=("false" "true")
DUAL_CONFIGS=("false_false" "true_false" "true_true")

WORKTREE_BASE="$SCRIPT_DIR/.eval_worktrees"
OUTPUT_DIR="$SCRIPT_DIR/.eval_results"
REPORT="$OUTPUT_DIR/comparison_report.txt"

mkdir -p "$OUTPUT_DIR"
rm -f "$OUTPUT_DIR"/*.csv "$OUTPUT_DIR"/*.txt 2>/dev/null || true

extract_metric() {
    local line="$1"
    local key="$2"
    echo "$line" | tr '|' '\n' | sed -n "s/^${key}=//p" | head -n 1
}

# ---- Cleanup on exit ----
cleanup() {
    echo ""
    echo "Cleaning up worktrees..."
    for entry in "${COMMITS[@]}"; do
        commit="${entry%%:*}"
        wt_path="$WORKTREE_BASE/$commit"
        if [ -d "$wt_path" ]; then
            git worktree remove --force "$wt_path" 2>/dev/null || true
        fi
    done
    rm -rf "$WORKTREE_BASE"
}
trap cleanup EXIT

echo "=========================================="
echo "Range Algorithm Evaluation - Commit Matrix"
echo "=========================================="
echo ""

mkdir -p "$WORKTREE_BASE"

for entry in "${COMMITS[@]}"; do
    commit="${entry%%:*}"
    label="${entry##*:}"
    wt_path="$WORKTREE_BASE/$commit"
    outfile="$OUTPUT_DIR/${commit}_eval.txt"
    short="${commit:0:8}"

    echo ">> Setting up worktree for $short ($label)..."
    if ! git worktree add "$wt_path" "$commit" 2>/dev/null; then
        echo "   ERROR: could not create worktree for $short" >&2
        continue
    fi

    # Inject current eval test file into the worktree.
    cp "$SCRIPT_DIR/range_eval_test.go" "$wt_path/"

    {
        printf "=== COMMIT %s (%s) ===\n" "$short" "$label"
        date
        echo ""
    } > "$outfile"

    echo "   Running TestRangeEvalMatrix..."
    if (cd "$wt_path" && go test -v -run '^TestRangeEvalMatrix$' -count=1 2>&1) >> "$outfile"; then
        echo "   OK"
    else
        echo "   FAILED (see $outfile)" >&2
    fi

    echo "   Collecting CSV file..."
    src="$wt_path/range_eval_matrix.csv"
    if [ -f "$src" ]; then
        cp "$src" "$OUTPUT_DIR/${short}_range_eval_matrix.csv"
        echo "   Saved: ${short}_range_eval_matrix.csv"
    else
        echo "   (no range_eval_matrix.csv produced)"
    fi

    echo ""
done

# ---- Generate summary report ----
echo "=========================================="
echo "Generating comparison_report.txt ..."
echo "=========================================="
echo ""

{
    printf "Range Algorithm Comparison Report\n"
    printf "Generated: %s\n" "$(date)"
    printf "\n"
    printf "Commits evaluated:\n"
    for entry in "${COMMITS[@]}"; do
        commit="${entry%%:*}"
        label="${entry##*:}"
        printf "  %s  %s\n" "${commit:0:8}" "$label"
    done
    printf "\n"
    printf "Matrix configs:\n"
    printf "  Single: false, true\n"
    printf "  Dual:   false_false, true_false, true_true\n"
    printf "Metrics: T0/T1/T2, label bucket rates (<5, 5-10, >10), coverage misses, pad/tightness, alignment, pad warn/bad, left/right axis splits\n"
    printf "\n"
    printf "%s\n" "$(printf '=%.0s' {1..110})"

    for entry in "${COMMITS[@]}"; do
        commit="${entry%%:*}"
        label="${entry##*:}"
        short="${commit:0:8}"
        outfile="$OUTPUT_DIR/${commit}_eval.txt"

        printf "\n### COMMIT %s (%s)\n\n" "$short" "$label"

        if [ ! -f "$outfile" ]; then
            printf "  (no output file found)\n"
            continue
        fi

        printf "  [Catalog]\n"
        if grep -q "EVAL|catalog|" "$outfile"; then
            grep "EVAL|catalog|" "$outfile" | sed 's/^/    /'
        else
            printf "    (not found)\n"
        fi
        printf "\n"

        printf "  [Single-Axis]\n"
        if grep -q "EVAL|single_summary|" "$outfile"; then
            grep "EVAL|single_summary|" "$outfile" | sed 's/^/    /'
        else
            printf "    (not found)\n"
        fi
        printf "\n"

        printf "  [Dual-Axis]\n"
        if grep -q "EVAL|dual_summary|" "$outfile"; then
            grep "EVAL|dual_summary|" "$outfile" | sed 's/^/    /'
        else
            printf "    (not found)\n"
        fi
        printf "\n"

        printf "%s\n" "$(printf -- '-%.0s' {1..110})"
    done

    printf "\n%s\n" "$(printf '=%.0s' {1..110})"
    printf "Cross-Commit: Single Config Metrics\n"
    printf "%s\n\n" "$(printf '=%.0s' {1..110})"
    printf "%-10s %-20s %-8s %-7s %-9s %-11s %-10s %-11s %-10s %-10s %-10s %-14s\n" "Commit" "Label" "Config" "T0Rate" "LblLT5Rate" "Lbl5to10Rate" "LblGT10Rate" "PadWarnRate" "PadBadRate" "PadAvg" "TightAvg" "CoverageMissRate"
    printf "%s\n" "$(printf -- '-%.0s' {1..110})"

    for entry in "${COMMITS[@]}"; do
        commit="${entry%%:*}"
        label="${entry##*:}"
        short="${commit:0:8}"
        outfile="$OUTPUT_DIR/${commit}_eval.txt"
        [ -f "$outfile" ] || continue

        for cfg in "${SINGLE_CONFIGS[@]}"; do
            line=$(grep "EVAL|single_summary|Config=${cfg}|" "$outfile" | head -n 1 || true)
            if [ -z "$line" ]; then
                printf "%-10s %-20s %-8s %-7s %-9s %-11s %-10s %-11s %-10s %-10s %-10s %-14s\n" "$short" "$label" "$cfg" "n/a" "n/a" "n/a" "n/a" "n/a" "n/a" "n/a" "n/a" "n/a"
                continue
            fi
            t0rate=$(extract_metric "$line" "T0Rate")
            llt5=$(extract_metric "$line" "LblLT5Rate")
            lrate=$(extract_metric "$line" "Lbl5to10Rate")
            lgt10=$(extract_metric "$line" "LblGT10Rate")
            padwarn=$(extract_metric "$line" "PadWarnRate")
            padbad=$(extract_metric "$line" "PadBadRate")
            padavg=$(extract_metric "$line" "PadAvg")
            tightavg=$(extract_metric "$line" "TightAvg")
            miss=$(extract_metric "$line" "CoverageMissRate")
            printf "%-10s %-20s %-8s %-7s %-9s %-11s %-10s %-11s %-10s %-10s %-10s %-14s\n" "$short" "$label" "$cfg" "$t0rate" "$llt5" "$lrate" "$lgt10" "$padwarn" "$padbad" "$padavg" "$tightavg" "$miss"
        done
    done

    printf "\n%s\n" "$(printf '=%.0s' {1..110})"
    printf "Cross-Commit: Dual Config Metrics\n"
    printf "%s\n\n" "$(printf '=%.0s' {1..110})"
    printf "%-10s %-20s %-12s %-9s %-7s %-9s %-11s %-10s %-11s %-10s %-10s %-10s %-14s %-10s %-11s %-10s %-10s %-11s %-10s\n" "Commit" "Label" "Config" "AlignRate" "T0Rate" "LblLT5Rate" "Lbl5to10Rate" "LblGT10Rate" "PadWarnRate" "PadBadRate" "PadAvg" "TightAvg" "CoverageMissRate" "LT0Rate" "LLbl5to10" "LCovMiss" "RT0Rate" "RLbl5to10" "RCovMiss"
    printf "%s\n" "$(printf -- '-%.0s' {1..110})"

    for entry in "${COMMITS[@]}"; do
        commit="${entry%%:*}"
        label="${entry##*:}"
        short="${commit:0:8}"
        outfile="$OUTPUT_DIR/${commit}_eval.txt"
        [ -f "$outfile" ] || continue

        for cfg in "${DUAL_CONFIGS[@]}"; do
            line=$(grep "EVAL|dual_summary|Config=${cfg}|" "$outfile" | head -n 1 || true)
            if [ -z "$line" ]; then
                printf "%-10s %-20s %-12s %-9s %-7s %-9s %-11s %-10s %-11s %-10s %-10s %-10s %-14s %-10s %-11s %-10s %-10s %-11s %-10s\n" "$short" "$label" "$cfg" "n/a" "n/a" "n/a" "n/a" "n/a" "n/a" "n/a" "n/a" "n/a" "n/a" "n/a" "n/a" "n/a" "n/a" "n/a" "n/a"
                continue
            fi
            align=$(extract_metric "$line" "AlignRate")
            t0rate=$(extract_metric "$line" "T0Rate")
            llt5=$(extract_metric "$line" "LblLT5Rate")
            lrate=$(extract_metric "$line" "Lbl5to10Rate")
            lgt10=$(extract_metric "$line" "LblGT10Rate")
            padwarn=$(extract_metric "$line" "PadWarnRate")
            padbad=$(extract_metric "$line" "PadBadRate")
            padavg=$(extract_metric "$line" "PadAvg")
            tightavg=$(extract_metric "$line" "TightAvg")
            miss=$(extract_metric "$line" "CoverageMissRate")
            lt0rate=$(extract_metric "$line" "LeftT0Rate")
            llbl5to10=$(extract_metric "$line" "LeftLbl5to10Rate")
            lmiss=$(extract_metric "$line" "LeftCoverageMissRate")
            rt0rate=$(extract_metric "$line" "RightT0Rate")
            rlbl5to10=$(extract_metric "$line" "RightLbl5to10Rate")
            rmiss=$(extract_metric "$line" "RightCoverageMissRate")
            printf "%-10s %-20s %-12s %-9s %-7s %-9s %-11s %-10s %-11s %-10s %-10s %-10s %-14s %-10s %-11s %-10s %-10s %-11s %-10s\n" "$short" "$label" "$cfg" "$align" "$t0rate" "$llt5" "$lrate" "$lgt10" "$padwarn" "$padbad" "$padavg" "$tightavg" "$miss" "$lt0rate" "$llbl5to10" "$lmiss" "$rt0rate" "$rlbl5to10" "$rmiss"
        done
    done

    printf "\n%s\n" "$(printf '=%.0s' {1..110})"
    printf "CSV File Row Counts\n"
    printf "%s\n\n" "$(printf '=%.0s' {1..110})"
    printf "%-45s  %s\n" "File" "Rows (incl. header)"
    printf "%s\n" "$(printf -- '-%.0s' {1..70})"
    for csv in "$OUTPUT_DIR"/*.csv; do
        [ -f "$csv" ] || continue
        rows=$(wc -l < "$csv" 2>/dev/null || echo "?")
        printf "%-45s  %s\n" "$(basename "$csv")" "$rows"
    done

} > "$REPORT"

cat "$REPORT"

# Clean up CSV files after report generation
rm -f "$OUTPUT_DIR"/*.csv

echo ""
echo "=========================================="
echo "Full per-commit output in: $OUTPUT_DIR"
printf "  %s\n" "$OUTPUT_DIR"/*.txt
echo "=========================================="
