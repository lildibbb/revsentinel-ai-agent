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
