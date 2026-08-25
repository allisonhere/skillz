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
