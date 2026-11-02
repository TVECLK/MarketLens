package ingestion

import (
	"fmt"
	"strings"
)

// Schema is the static JSON schema for LLM extraction. It never depends
// on DB data, so it's a package-level value built once rather than
// rebuilt on every call.
var Schema = map[string]any{
	"type":                 "object",
	"additionalProperties": false,
	"properties": map[string]any{
		"job_type": map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"type": map[string]any{"type": "string"},
			},
			"required": []string{"type"},
		},
		"work_mode": map[string]any{
			"type": "string",
			"enum": []string{"remote", "onsite", "hybrid"},
		},
		"no_of_vacancies": map[string]any{"type": "integer", "minimum": 1},
		"meta_data": map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties": map[string]any{
				"geo_data": map[string]any{
					"type":                 "object",
					"additionalProperties": false,
					"properties": map[string]any{
						"province": map[string]any{"type": "string"},
					},
					"required": []string{"province"},
				},
				"posted_at": map[string]any{
					"type":    "string",
					"pattern": `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`,
				},
				"confidence_score":        map[string]any{"type": "number", "minimum": 0, "maximum": 1},
				"formality_id":            map[string]any{"type": "integer"},
				"gender_id":               map[string]any{"type": "integer"},
				"vocational_education_id": map[string]any{"type": "integer"},
				"employment_sector_id":    map[string]any{"type": "integer"},
				"education_level_id":      map[string]any{"type": "integer"},
				"experience_id":           map[string]any{"type": "integer"},
			},
			"required": []string{
				"geo_data", "confidence_score",
				"formality_id", "gender_id", "vocational_education_id",
				"employment_sector_id", "education_level_id", "experience_id",
			},
		},
		"skills": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties": map[string]any{
					"skill": map[string]any{"type": "string"},
				},
				"required": []string{"skill"},
			},
		},
	},
	"required": []string{
		"job_type", "work_mode", "no_of_vacancies", "meta_data", "skills",
	},
}

const instructionTemplate = `Extract structural data from the raw job text into the specified JSON format.
CRITICAL: If a field is missing, use the strict default provided. Do not hallucinate.

Fields & Defaults:

1. 'job_type': Object with contract type.
Format: {"type": "Full Time"} or {"type": "Part Time"} or {"type": "Contract"} or {"type": "Internship"}
Default: {"type": "Full Time"}

2. 'no_of_vacancies': Number of open positions for this role, if stated.
Default: 1

3. 'work_mode': The job's work arrangement.
Must be one of: "remote", "onsite", "hybrid"
- "remote": explicit remote/work-from-home wording exists, with no on-site requirement
- "hybrid": explicit mention of a mix of remote and in-office work
- "onsite": no remote/hybrid wording, or the role explicitly requires in-office presence
Default: "onsite"

4. 'meta_data': Object containing all metadata fields:

- 'geo_data': Infer the Sri Lankan province from the location text.
    Format: {"province": "<province>"}
    Must match EXACTLY ONE from this list:
    ["Western", "Central", "Northern", "Eastern", "North Western",
        "North Central", "Uva", "Southern", "Sabaragamuwa"]
    Default: {"province": "Western"}

- 'posted_at': ISO 8601 timestamp, ONLY if an explicit posting date is stated in the job text (e.g. "Posted on 15th March 2026").
    Format: "YYYY-MM-DDTHH:MM:SSZ"
    You do not know the current date — if no explicit date is stated, OMIT this field rather than guessing. The system fills in today's date automatically when it's absent.

- 'confidence_score': Your confidence in the extraction accuracy.
    Float between 0.00 and 1.00. Use 1.00 if all fields are clearly present.
    Default: 0.80

- 'formality_id': Must be one of these ids:
%s

- 'gender_id': Must be one of these ids:
%s

- 'vocational_education_id': Must be one of these ids:
%s

- 'employment_sector_id': Must be one of these ids:
%s

- 'education_level_id': Must be one of these ids:
%s

- 'experience_id': Must be one of these ids:
%s

5. 'skills': Array of skill objects extracted from the job description.
Format: [{"skill": "Skill Name"}, ...]
Extract only concrete technical or professional skills mentioned.
Default: []
`

// buildExtractionPrompt assembles the final prompt sent to the LLM for
// job extraction: the filled-in instruction text, the JSON Schema the
// response must conform to, and the raw job text to extract from.
func buildExtractionPrompt(instruction string, schemaJSON []byte, jobText string) string {
	return instruction +
		"\n\nThe JSON you return MUST conform to this JSON Schema:\n" + string(schemaJSON) +
		"\n\nJob posting to extract from:\n" + jobText
}

// buildAskToPickPrompt assembles the prompt sent to the LLM for a single
// classification step: pick the best-matching option at one hierarchy
// level, or reject all of them.
func buildAskToPickPrompt(jobText string, options []Option, levelName string) string {
	var optionLines strings.Builder
	for _, o := range options {
		fmt.Fprintf(&optionLines, "%d: %s\n", o.ID, o.Name)
	}

	return fmt.Sprintf(
		"You are classifying a job posting into a %s.\n\n"+
			"Job details:\n%s\n\n"+
			"Choose the single best matching %s from this list:\n%s\n"+
			"IMPORTANT: The id you return MUST be exactly one of the ids "+
			"listed above — do not invent, guess, or modify an id. If none "+
			"of the options are a good fit, choose the closest match rather "+
			"than fabricating a new id.\n"+
			`Respond ONLY with JSON in this exact format: {"id": <chosen id as integer>}`,
		levelName, jobText, levelName, optionLines.String(),
	)
}
