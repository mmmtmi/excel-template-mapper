# Lesson 2 Answer

1. `name`
2. `input.Name`
3. `InsertTemplate`

完成形:

```go
if strings.TrimSpace(input.Name) == "" {
  return nil, errors.New("name is required")
}

tpl := &internalmodel.Template{
  ID:           uuid.NewString(),
  Name:         input.Name,
  Target:       input.Name,
  SheetName:    input.SheetName,
  HeaderRow:    int(input.HeaderRow),
  DataStartRow: int(input.DataStartRow),
}

if err := mysql.InsertTemplate(ctx, r.DB, tpl, now); err != nil {
  return nil, err
}
```
