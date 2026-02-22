const API_URL = import.meta.env.VITE_GRAPHQL_ENDPOINT || "http://localhost:8080/query";

async function parseResponse(res) {
  const payload = await res.json();
  if (!res.ok) {
    throw new Error(payload?.errors?.[0]?.message || `Request failed: ${res.status}`);
  }
  if (payload.errors?.length) {
    throw new Error(payload.errors[0].message);
  }
  return payload.data;
}

export async function gqlRequest(query, variables = {}) {
  const res = await fetch(API_URL, {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify({ query, variables })
  });
  return parseResponse(res);
}

export async function gqlUpload({ query, variables, fileVarName, file }) {
  const operations = JSON.stringify({
    query,
    variables: { ...variables, [fileVarName]: null }
  });
  const map = JSON.stringify({ "0": [`variables.${fileVarName}`] });

  const formData = new FormData();
  formData.append("operations", operations);
  formData.append("map", map);
  formData.append("0", file);

  const res = await fetch(API_URL, {
    method: "POST",
    body: formData
  });

  return parseResponse(res);
}

