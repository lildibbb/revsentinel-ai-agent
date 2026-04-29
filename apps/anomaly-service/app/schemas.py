from typing import Literal
from uuid import UUID

from pydantic import BaseModel, ConfigDict, Field


class ScoreRequest(BaseModel):
    feature_vector: dict


class GenerateReasoningRequest(BaseModel):
    model_config = ConfigDict(extra="forbid")

    case_id: UUID
    force_regenerate: bool = False


class Citation(BaseModel):
    model_config = ConfigDict(extra="forbid")

    source: str = Field(min_length=1)
    reference: str = Field(min_length=1)
    excerpt: str = Field(min_length=1)


class GenerateReasoningResponse(BaseModel):
    model_config = ConfigDict(extra="forbid")

    status: Literal["success", "failed"]
    prompt_version: str = Field(min_length=1)
    response_schema_version: str = Field(min_length=1)
    summary: str = Field(min_length=1)
    recommended_action: str = Field(min_length=1)
    citations: list[Citation]
    amount_original: float
    currency_original: str = Field(min_length=3, max_length=3)
    amount_myr_normalized: float
    fx_rate_to_myr: float
