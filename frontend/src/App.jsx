import { useEffect, useMemo, useState } from "react";
import { gqlRequest, gqlUpload } from "./graphql";

const DRAFT_FROM_EXCEL = `
  mutation DraftRulesFromExcel($file: Upload!) {
    draftRulesFromExcel(file: $file) {
      headers
      sampleRows
      draftRules {
        sourceType
        sourceKey
        targetLabel
        transform
        required
      }
    }
  }
`;

const PROCESS_EXCEL_ONLY = `
  mutation ProcessExcelOnly($file: Upload!) {
    processExcelOnly(file: $file)
  }
`;

const PROCESS_EXCEL = `
  mutation ProcessExcel($templateName: String!, $file: Upload!) {
    processExcel(templateName: $templateName, file: $file)
  }
`;

const CREATE_TEMPLATE = `
  mutation CreateTemplate($input: CreateTemplateInput!) {
    createTemplate(input: $input) {
      id
      name
      target
      sheetName
      headerRow
      dataStartRow
    }
  }
`;

const CREATE_RULE = `
  mutation CreateRule($input: CreateRuleInput!) {
    createRule(input: $input) {
      id
      templateId
      sourceType
      sourceKey
      targetLabel
      transform
      required
    }
  }
`;

const LIST_TEMPLATES = `
  query Templates {
    templates {
      id
      name
      target
      sheetName
      headerRow
      dataStartRow
    }
  }
`;

const GET_TEMPLATE = `
  query Template($name: String!) {
    template(name: $name) {
      id
      name
      target
      sheetName
      headerRow
      dataStartRow
    }
  }
`;

const GET_RULES = `
  query Rules($templateId: ID!) {
    rules(templateId: $templateId) {
      id
      templateId
      sourceType
      sourceKey
      targetLabel
      transform
      required
      priority
      createdAt
      updatedAt
    }
  }
`;

const UPDATE_TEMPLATE = `
  mutation UpdateTemplate($id: ID!, $input: UpdateTemplateInput!) {
    updateTemplate(id: $id, input: $input) {
      id
      name
      target
      sheetName
      headerRow
      dataStartRow
    }
  }
`;

const UPDATE_RULE = `
  mutation UpdateRule($id: ID!, $input: UpdateRuleInput!) {
    updateRule(id: $id, input: $input) {
      id
      templateId
      sourceType
      sourceKey
      targetLabel
      transform
      required
      priority
      createdAt
      updatedAt
    }
  }
`;

function toPrettyJSON(value) {
  if (!value) return "";
  try {
    const parsed = JSON.parse(value);
    return JSON.stringify(parsed, null, 2);
  } catch (_) {
    return value;
  }
}

export default function App() {
  const [draftFile, setDraftFile] = useState(null);
  const [processOnlyFile, setProcessOnlyFile] = useState(null);
  const [processTemplateFile, setProcessTemplateFile] = useState(null);
  const [templateName, setTemplateName] = useState("");
  const [createTemplateName, setCreateTemplateName] = useState("");
  const [sheetName, setSheetName] = useState("");
  const [headerRow, setHeaderRow] = useState(1);
  const [dataStartRow, setDataStartRow] = useState(2);
  const [draftPayload, setDraftPayload] = useState(null);
  const [excelOnlyJSON, setExcelOnlyJSON] = useState("");
  const [templateJSON, setTemplateJSON] = useState("");
  const [createdTemplate, setCreatedTemplate] = useState(null);
  const [templates, setTemplates] = useState([]);
  const [selectedTemplate, setSelectedTemplate] = useState(null);
  const [selectedRules, setSelectedRules] = useState([]);
  const [isBusy, setIsBusy] = useState(false);
  const [error, setError] = useState("");

  const rules = useMemo(() => draftPayload?.draftRules || [], [draftPayload]);

  const setRuleField = (index, key, value) => {
    setDraftPayload((prev) => {
      if (!prev) return prev;
      const nextRules = prev.draftRules.map((rule, i) => {
        if (i !== index) return rule;
        return { ...rule, [key]: value };
      });
      return { ...prev, draftRules: nextRules };
    });
  };

  const runWithGuard = async (task) => {
    setError("");
    setIsBusy(true);
    try {
      await task();
    } catch (e) {
      setError(e.message || "Unknown error");
    } finally {
      setIsBusy(false);
    }
  };

  const loadTemplates = async () => {
    const data = await gqlRequest(LIST_TEMPLATES);
    setTemplates(data.templates || []);
  };

  const loadTemplateDetails = async (name) => {
    if (!name) {
      setSelectedTemplate(null);
      setSelectedRules([]);
      return;
    }
    const tData = await gqlRequest(GET_TEMPLATE, { name });
    const tpl = tData.template;
    setSelectedTemplate(tpl);
    const rData = await gqlRequest(GET_RULES, { templateId: tpl.id });
    setSelectedRules(rData.rules || []);
  };

  useEffect(() => {
    runWithGuard(loadTemplates);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleDraftFromExcel = () =>
    runWithGuard(async () => {
      if (!draftFile) throw new Error("Pick an Excel file for draft preview");
      const data = await gqlUpload({
        query: DRAFT_FROM_EXCEL,
        variables: {},
        fileVarName: "file",
        file: draftFile
      });
      setDraftPayload(data.draftRulesFromExcel);
      setCreatedTemplate(null);
    });

  const handleProcessOnly = () =>
    runWithGuard(async () => {
      if (!processOnlyFile) throw new Error("Pick an Excel file for ExcelOnly");
      const data = await gqlUpload({
        query: PROCESS_EXCEL_ONLY,
        variables: {},
        fileVarName: "file",
        file: processOnlyFile
      });
      setExcelOnlyJSON(toPrettyJSON(data.processExcelOnly));
    });

  const handleProcessWithTemplate = () =>
    runWithGuard(async () => {
      if (!templateName.trim()) throw new Error("Template name is required");
      if (!processTemplateFile) throw new Error("Pick an Excel file for template process");
      const data = await gqlUpload({
        query: PROCESS_EXCEL,
        variables: { templateName: templateName.trim() },
        fileVarName: "file",
        file: processTemplateFile
      });
      setTemplateJSON(toPrettyJSON(data.processExcel));
    });

  const handleSaveTemplateAndRules = () =>
    runWithGuard(async () => {
      if (!createTemplateName.trim()) throw new Error("Template name is required");
      if (!draftPayload?.draftRules?.length) throw new Error("No draft rules to save");

      const tplData = await gqlRequest(CREATE_TEMPLATE, {
        input: {
          name: createTemplateName.trim(),
          sheetName: sheetName.trim() ? sheetName.trim() : null,
          headerRow: Number(headerRow),
          dataStartRow: Number(dataStartRow)
        }
      });

      const tpl = tplData.createTemplate;
      for (const draftRule of draftPayload.draftRules) {
        await gqlRequest(CREATE_RULE, {
          input: {
            templateId: tpl.id,
            sourceType: draftRule.sourceType,
            sourceKey: draftRule.sourceKey,
            targetLabel: draftRule.targetLabel,
            transform: draftRule.transform || null,
            required: Boolean(draftRule.required)
          }
        });
      }
      setCreatedTemplate(tpl);
      setTemplateName(tpl.name);
      await loadTemplates();
    });

  const setSelectedTemplateField = (key, value) => {
    setSelectedTemplate((prev) => {
      if (!prev) return prev;
      return { ...prev, [key]: value };
    });
  };

  const setSelectedRuleField = (index, key, value) => {
    setSelectedRules((prev) =>
      prev.map((rule, i) => (i === index ? { ...rule, [key]: value } : rule))
    );
  };

  const handleUpdateTemplate = () =>
    runWithGuard(async () => {
      if (!selectedTemplate?.id) throw new Error("Select a template first");
      await gqlRequest(UPDATE_TEMPLATE, {
        id: selectedTemplate.id,
        input: {
          name: selectedTemplate.name,
          sheetName: selectedTemplate.sheetName || null,
          headerRow: Number(selectedTemplate.headerRow),
          dataStartRow: Number(selectedTemplate.dataStartRow)
        }
      });
      await loadTemplates();
      await loadTemplateDetails(selectedTemplate.name);
      setTemplateName(selectedTemplate.name);
    });

  const handleUpdateRule = (index) =>
    runWithGuard(async () => {
      const rule = selectedRules[index];
      if (!rule?.id) throw new Error("Rule id missing");
      await gqlRequest(UPDATE_RULE, {
        id: rule.id,
        input: {
          sourceType: rule.sourceType,
          sourceKey: rule.sourceKey,
          targetLabel: rule.targetLabel,
          // Send "" when empty so backend can treat it as "clear transform".
          transform: rule.transform ?? "",
          required: Boolean(rule.required)
        }
      });
      if (selectedTemplate?.id) {
        const rData = await gqlRequest(GET_RULES, { templateId: selectedTemplate.id });
        setSelectedRules(rData.rules || []);
      }
    });

  return (
    <div className="page">
      <div className="bg-orb bg-orb-a" />
      <div className="bg-orb bg-orb-b" />
      <main className="layout">
        <header className="hero">
          <p className="eyebrow">Excel Template Mapper</p>
          <h1>React Frontend for Upload + Rule Draft + Process</h1>
          <p>
            Flow: upload Excel for draft preview, edit rules, save template/rules, then run template
            process.
          </p>
        </header>

        {error && <div className="error">{error}</div>}

        <section className="panel">
          <h2>1) Draft Rules From Excel</h2>
          <div className="row">
            <input
              type="file"
              accept=".xlsx"
              onChange={(e) => setDraftFile(e.target.files?.[0] || null)}
            />
            <button onClick={handleDraftFromExcel} disabled={isBusy}>
              {isBusy ? "Running..." : "Generate Draft"}
            </button>
          </div>

          {draftPayload && (
            <>
              <div className="meta">
                <span>Headers: {draftPayload.headers.length}</span>
                <span>Sample rows: {draftPayload.sampleRows.length}</span>
              </div>
              <div className="sample-wrap">
                <table>
                  <thead>
                    <tr>
                      {draftPayload.headers.map((header) => (
                        <th key={header}>{header}</th>
                      ))}
                    </tr>
                  </thead>
                  <tbody>
                    {draftPayload.sampleRows.map((row, rowIndex) => (
                      <tr key={`row-${rowIndex}`}>
                        {row.map((cell, colIndex) => (
                          <td key={`cell-${rowIndex}-${colIndex}`}>{cell}</td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              <div className="rule-list">
                {rules.map((rule, index) => (
                  <div className="rule-row" key={`${rule.sourceKey}-${index}`}>
                    <input
                      value={rule.sourceKey}
                      onChange={(e) => setRuleField(index, "sourceKey", e.target.value)}
                    />
                    <input
                      value={rule.targetLabel}
                      onChange={(e) => setRuleField(index, "targetLabel", e.target.value)}
                    />
                    <select
                      value={rule.sourceType}
                      onChange={(e) => setRuleField(index, "sourceType", e.target.value)}
                    >
                      <option value="HEADER">HEADER</option>
                      <option value="CELL">CELL</option>
                    </select>
                    <input
                      placeholder="transform (e.g. trim)"
                      value={rule.transform || ""}
                      onChange={(e) =>
                        setRuleField(index, "transform", e.target.value ? e.target.value : null)
                      }
                    />
                    <label className="check">
                      <input
                        type="checkbox"
                        checked={rule.required}
                        onChange={(e) => setRuleField(index, "required", e.target.checked)}
                      />
                      required
                    </label>
                  </div>
                ))}
              </div>
            </>
          )}
        </section>

        <section className="panel">
          <h2>2) Save Template + Rules</h2>
          <div className="grid">
            <input
              placeholder="template name"
              value={createTemplateName}
              onChange={(e) => setCreateTemplateName(e.target.value)}
            />
            <input
              placeholder="sheet name (optional)"
              value={sheetName}
              onChange={(e) => setSheetName(e.target.value)}
            />
            <input
              type="number"
              min="1"
              value={headerRow}
              onChange={(e) => setHeaderRow(e.target.value)}
            />
            <input
              type="number"
              min="1"
              value={dataStartRow}
              onChange={(e) => setDataStartRow(e.target.value)}
            />
          </div>
          <button onClick={handleSaveTemplateAndRules} disabled={isBusy}>
            {isBusy ? "Saving..." : "Save Draft to DB"}
          </button>
          {createdTemplate && (
            <p className="ok">
              Saved template: <strong>{createdTemplate.name}</strong> (id: {createdTemplate.id})
            </p>
          )}
        </section>

        <section className="panel">
          <h2>3) Process Excel Only</h2>
          <div className="row">
            <input
              type="file"
              accept=".xlsx"
              onChange={(e) => setProcessOnlyFile(e.target.files?.[0] || null)}
            />
            <button onClick={handleProcessOnly} disabled={isBusy}>
              {isBusy ? "Running..." : "Run ExcelOnly"}
            </button>
          </div>
          <textarea readOnly value={excelOnlyJSON} placeholder="Result JSON will appear here" />
        </section>

        <section className="panel">
          <h2>4) Process With Template</h2>
          <div className="row">
            <select
              value={templateName}
              onChange={(e) => {
                const next = e.target.value;
                setTemplateName(next);
                runWithGuard(() => loadTemplateDetails(next));
              }}
            >
              <option value="">Select template</option>
              {templates.map((tpl) => (
                <option key={tpl.id} value={tpl.name}>
                  {tpl.name}
                </option>
              ))}
            </select>
            <input
              type="file"
              accept=".xlsx"
              onChange={(e) => setProcessTemplateFile(e.target.files?.[0] || null)}
            />
            <button onClick={() => runWithGuard(loadTemplates)} disabled={isBusy}>
              {isBusy ? "Loading..." : "Reload Templates"}
            </button>
            <button onClick={handleProcessWithTemplate} disabled={isBusy}>
              {isBusy ? "Running..." : "Run Template Process"}
            </button>
          </div>
          <textarea readOnly value={templateJSON} placeholder="Result JSON will appear here" />
        </section>

        <section className="panel">
          <h2>5) Template Inspector (View/Update)</h2>
          <div className="row">
            <select
              value={selectedTemplate?.name || ""}
              onChange={(e) => runWithGuard(() => loadTemplateDetails(e.target.value))}
            >
              <option value="">Select template</option>
              {templates.map((tpl) => (
                <option key={tpl.id} value={tpl.name}>
                  {tpl.name}
                </option>
              ))}
            </select>
            <button onClick={() => runWithGuard(loadTemplates)} disabled={isBusy}>
              {isBusy ? "Loading..." : "Reload Templates"}
            </button>
          </div>

          {selectedTemplate && (
            <>
              <div className="grid">
                <input
                  placeholder="name"
                  value={selectedTemplate.name || ""}
                  onChange={(e) => setSelectedTemplateField("name", e.target.value)}
                />
                <input
                  placeholder="sheet name (optional)"
                  value={selectedTemplate.sheetName || ""}
                  onChange={(e) => setSelectedTemplateField("sheetName", e.target.value)}
                />
                <input
                  type="number"
                  min="1"
                  value={selectedTemplate.headerRow}
                  onChange={(e) => setSelectedTemplateField("headerRow", e.target.value)}
                />
                <input
                  type="number"
                  min="1"
                  value={selectedTemplate.dataStartRow}
                  onChange={(e) => setSelectedTemplateField("dataStartRow", e.target.value)}
                />
              </div>
              <button onClick={handleUpdateTemplate} disabled={isBusy}>
                {isBusy ? "Saving..." : "Update Template"}
              </button>

              <div className="rule-list">
                {selectedRules.map((rule, index) => (
                  <div className="rule-row" key={rule.id}>
                    <input
                      value={rule.sourceKey}
                      onChange={(e) => setSelectedRuleField(index, "sourceKey", e.target.value)}
                    />
                    <input
                      value={rule.targetLabel}
                      onChange={(e) => setSelectedRuleField(index, "targetLabel", e.target.value)}
                    />
                    <select
                      value={rule.sourceType}
                      onChange={(e) => setSelectedRuleField(index, "sourceType", e.target.value)}
                    >
                      <option value="HEADER">HEADER</option>
                      <option value="CELL">CELL</option>
                    </select>
                    <input
                      placeholder="transform (e.g. trim)"
                      value={rule.transform || ""}
                      onChange={(e) =>
                        setSelectedRuleField(index, "transform", e.target.value ? e.target.value : null)
                      }
                    />
                    <label className="check">
                      <input
                        type="checkbox"
                        checked={Boolean(rule.required)}
                        onChange={(e) => setSelectedRuleField(index, "required", e.target.checked)}
                      />
                      required
                    </label>
                    <button onClick={() => handleUpdateRule(index)} disabled={isBusy}>
                      {isBusy ? "..." : "Update Rule"}
                    </button>
                  </div>
                ))}
              </div>
            </>
          )}
        </section>
      </main>
    </div>
  );
}
