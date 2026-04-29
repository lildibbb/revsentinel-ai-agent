from fastapi.testclient import TestClient

from app.main import app


def test_generate_reasoning_success_contract():
    client = TestClient(app)
    res = client.post(
        "/reasoning/generate",
        json={"case_id": "00000000-0000-0000-0000-000000000001"},
    )

    assert res.status_code == 200
    body = res.json()
    assert body["status"] in {"success", "failed"}
    assert body["prompt_version"] == "v1"
    assert body["response_schema_version"] == "v1"
    assert "summary" in body
    assert "recommended_action" in body
    assert isinstance(body["citations"], list)
    assert "amount_original" in body
    assert "currency_original" in body
    assert "amount_myr_normalized" in body
    assert "fx_rate_to_myr" in body


def test_generate_reasoning_rejects_extra_fields():
    client = TestClient(app)
    res = client.post(
        "/reasoning/generate",
        json={
            "case_id": "00000000-0000-0000-0000-000000000001",
            "unexpected": "value",
        },
    )

    assert res.status_code == 422


def test_generate_reasoning_returns_typed_error_on_vertex_failure(monkeypatch):
    def _raise(*args, **kwargs):
        from app import vertex_client

        raise vertex_client.VertexCallError()

    monkeypatch.setattr("app.main.generate_grounded_reasoning", _raise)

    client = TestClient(app)
    res = client.post(
        "/reasoning/generate",
        json={"case_id": "00000000-0000-0000-0000-000000000001"},
    )

    assert res.status_code == 502
    body = res.json()
    assert body["status"] == "failed"
    assert body["error_code"] == "vertex_call_failed"


def test_generate_reasoning_returns_typed_error_on_retrieval_failure():
    client = TestClient(app)
    res = client.post(
        "/reasoning/generate",
        json={"case_id": "00000000-0000-0000-0000-000000000000"},
    )

    assert res.status_code == 502
    body = res.json()
    assert body["status"] == "failed"
    assert body["error_code"] == "retrieval_failed"
