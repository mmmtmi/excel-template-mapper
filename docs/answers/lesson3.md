# Lesson 3 Answer

1. `SourceKey`
2. `TargetLabel`
3. `outRow`

完成形:

```go
for _, r := range rules {
  if r.SourceType != "HEADER" {
    continue
  }

  val, err := applyTransform(row.Values[r.SourceKey], r.Transform)
  if err != nil {
    return nil, err
  }

  outRow[r.TargetLabel] = val
}

if err := validateRequiredValues(outRow, requiredRules, row.ExcelRow); err != nil {
  return nil, err
}
```
