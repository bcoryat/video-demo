package clarifai

// Service interface for Clarifai
type Service interface {
	PredictByBytes(timestamp int64, b64Str string, origB64Str string) (*FrameInfo, error)
}

type inputs struct {
	Inputs []dataByBytes `json:"inputs"`
}

type dataByBytes struct {
	ID   string `json:"id"`
	Data struct {
		Image struct {
			Base64 string `json:"base64"`
		} `json:"image"`
	} `json:"data"`
}

// PredictResponse contains most useful components of the clarifai predict API
type PredictResponse struct {
	Status struct {
		Code        int    `json:"code"`
		Description string `json:"description"`
	} `json:"status"`

	Outputs []struct {
		Input struct {
			ID   string `json:"id"`
			Data struct {
				Image struct {
					URL    string `json:"url"`
					Base64 string `json:"base64"`
				} `json:"image"`
			} `json:"data"`
		}
		Data struct {
			Regions []struct {
				ID         string `json:"id"`
				RegionInfo struct {
					BoundingBox struct {
						TopRow    float64 `json:"top_row"`
						LeftCol   float64 `json:"left_col"`
						BottomRow float64 `json:"bottom_row"`
						RightCol  float64 `json:"right_col"`
					} `json:"bounding_box"`
				} `json:"region_info"`
				RegionData struct {
					Concepts []struct {
						ID    string  `json:"id"`
						Name  string  `json:"name"`
						Value float64 `json:"value"`
					} `json:"concepts"`
				} `json:"data"`
			} `json:"regions"`
		} `json:"data"`
	} `json:"outputs"`
}

// FrameInfo is a front-end friendly transformed PredictResponse
type FrameInfo struct {
	InputID string   `json:"inputID"`
	URL     string   `json:"url"`
	B64     string   `json:"b64"`
	Objects []object `json:"objects"`
}
type object struct {
	BoundingBox  bBox    `json:"bbox"`
	ConceptName  string  `json:"concept_name"`
	ConceptValue float64 `json:"concept_value"`
}
type bBox struct {
	TopRow    float64 `json:"top_row"`
	LeftCol   float64 `json:"left_col"`
	BottomRow float64 `json:"bottom_row"`
	RightCol  float64 `json:"right_col"`
}
