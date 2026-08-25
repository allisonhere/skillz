# Compact Terminal Graphs

## Choose the Smallest Useful Encoding

A graph should answer a question faster than text. Do not add one merely because dashboards are expected to have charts.

## Sparklines

Use for recent trend direction where exact individual values are secondary.

`▁▂▃▅▄▆▇█`

Good for CPU, network, latency, queue depth, request rate, temperature, and memory trends.

Always pair with a current value when that value matters:

`CPU  42%  ▁▂▃▅▄▆▅▄`

## Segmented Meters

Use for capacity or completion:

`▰▰▰▰▰▱▱▱▱▱  50%`

or

`■■■■■□□□□□  5/10`

These work especially well in one-row resource summaries.

### Implementation

```go
var sparkRamp = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// Sparkline renders the last `width` samples. Scale is 0..max; when max <= 0 it
// autoscales to the window's peak. Missing samples (NaN) render as '·' so
// "no data" never looks like zero.
func Sparkline(samples []float64, width int, maxVal float64) string {
    if width <= 0 {
        return ""
    }
    if len(samples) > width {
        samples = samples[len(samples)-width:]
    }
    if maxVal <= 0 {
        for _, v := range samples {
            if !math.IsNaN(v) && v > maxVal {
                maxVal = v
            }
        }
    }
    var b strings.Builder
    for i := 0; i < width-len(samples); i++ {
        b.WriteRune('·') // pad history we don't have yet
    }
    for _, v := range samples {
        switch {
        case math.IsNaN(v):
            b.WriteRune('·')
        case maxVal <= 0:
            b.WriteRune(sparkRamp[0])
        default:
            idx := int(v / maxVal * float64(len(sparkRamp)-1))
            b.WriteRune(sparkRamp[min(max(idx, 0), len(sparkRamp)-1)])
        }
    }
    return b.String()
}

// Meter renders a fraction as `width` segmented cells. frac is clamped to 0..1;
// a negative frac means "unavailable" and renders as dots.
func Meter(frac float64, width int) string {
    if width <= 0 {
        return ""
    }
    if frac < 0 || math.IsNaN(frac) {
        return strings.Repeat("·", width)
    }
    frac = math.Min(frac, 1)
    filled := int(math.Round(frac * float64(width)))
    return strings.Repeat("▰", filled) + strings.Repeat("▱", width-filled)
}
```

Both return exactly `width` cells for every input — that is what keeps dashboard columns aligned. Assert it in a test.

For the ASCII Unicode level, swap ramps behind the same signature: `#` / `-` for meters, and `_.-~=` style ramps or a bare number for sparklines.

## Block Meters

Use full/partial block characters when terminal support is reliable and finer visual resolution matters.

Avoid complicated multi-color gradients unless the colors encode real thresholds.

## Thresholds

Do not imply universal health thresholds. If thresholds are meaningful, derive them from domain rules/configuration and expose warning/error states redundantly with glyph or text.

## History Windows

Label or document the time window when a sparkline could otherwise mislead. For example:

`NET  18 MB/s  ▁▁▂▄█▆▃▂  60s`

## Zero and Missing Data

Differentiate zero from unavailable data. A blank meter should not silently mean both.

Possible unavailable rendering:

`NET  —  ········`

## Width Adaptation

Shrink graph history before sacrificing essential labels and current values. On very narrow terminals, fall back to a numeric value plus status glyph.
