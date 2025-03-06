package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/xhd2015/data-driven-testing/decision_tree"
	"github.com/xhd2015/data-driven-testing/decision_tree/svg"
)

func main() {
	// Create a sample tree that matches the image
	tree := &decision_tree.Node{
		ID:    "root",
		Label: "Feature Status Check",
		Conditions: map[string]any{
			"feature_id": 8,
		},
		Children: []*decision_tree.Node{
			{
				ID:    "frozen",
				Label: "Feature Frozen",
				Children: []*decision_tree.Node{
					{
						ID:    "hide_feature",
						Label: "Hide Feature",
						Conditions: map[string]any{
							"user_state":    3,
							"feature_type":  1,
							"feature_state": 2,
						},
					},
				},
			},
			{
				ID:    "inactive",
				Label: "Feature Inactive",
				Children: []*decision_tree.Node{
					{
						ID:    "display_activation",
						Label: "Display Activation Entry",
						Conditions: map[string]any{
							"user_state":    1,
							"feature_type":  2,
							"feature_state": 4,
							"config_link":   "",
						},
					},
				},
			},
			{
				ID:    "activated",
				Label: "Feature Activated",
				Children: []*decision_tree.Node{
					{
						ID:    "limit_enough",
						Label: "Limit Sufficient",
					},
					{
						ID:    "limit_not_enough",
						Label: "Limit Insufficient",
						Children: []*decision_tree.Node{
							{
								ID:    "xtra_ineligible",
								Label: "Extension Ineligible",
								Children: []*decision_tree.Node{
									{
										ID:    "error_popup",
										Label: "Display Error Message",
										Conditions: map[string]any{
											"user_state":    2,
											"feature_type":  1,
											"feature_state": 1,
										},
									},
								},
							},
							{
								ID:    "xtra_eligible",
								Label: "Extension Eligible",
								Children: []*decision_tree.Node{
									{
										ID:    "ext_not_activated",
										Label: "Extension Not Activated",
										Children: []*decision_tree.Node{
											{
												ID:    "auto_trigger",
												Label: "Auto Trigger Activation",
												Conditions: map[string]any{
													"user_state":    2,
													"feature_type":  2,
													"feature_state": 4,
												},
											},
										},
									},
									{
										ID:    "ext_activated",
										Label: "Extension Activated",
										Children: []*decision_tree.Node{
											{
												ID:    "ext_limit_available",
												Label: "Extension Limit Available",
												Children: []*decision_tree.Node{
													{
														ID:    "ext_limit_enough",
														Label: "Extension Limit Sufficient",
														Children: []*decision_tree.Node{
															{
																ID:    "display_plans",
																Label: "Display Payment Plans",
																Conditions: map[string]any{
																	"user_state":       2,
																	"feature_type":     2,
																	"feature_state":    0,
																	"transaction_type": 7,
																},
															},
														},
													},
													{
														ID:    "ext_limit_not_enough",
														Label: "Extension Limit Insufficient",
														Children: []*decision_tree.Node{
															{
																ID:    "show_options",
																Label: "Show Payment Options",
																Conditions: map[string]any{
																	"user_state":    2,
																	"feature_type":  2,
																	"feature_state": 4,
																},
															},
														},
													},
												},
											},
											{
												ID:    "ext_frozen",
												Label: "Extension Frozen",
												Children: []*decision_tree.Node{
													{
														ID:    "ext_limit_expired",
														Label: "Extension Limit Expired",
														Children: []*decision_tree.Node{
															{
																ID:    "display_renewal",
																Label: "Show Renewal Options",
																Conditions: map[string]any{
																	"user_state":    2,
																	"feature_type":  2,
																	"feature_state": 5,
																},
															},
														},
													},
													{
														ID:    "others",
														Label: "Other Cases",
														Children: []*decision_tree.Node{
															{
																ID:    "show_current",
																Label: "Show Current Status",
																Conditions: map[string]any{
																	"user_state":    2,
																	"feature_type":  2,
																	"feature_state": 7,
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Create renderer
	renderer := svg.NewRenderer(decision_tree.DefaultConfig())

	// Generate centered layout (default)
	svgContent := renderer.RenderTree(tree)
	err := os.WriteFile("tree.svg", []byte(svgContent), 0644)
	if err != nil {
		log.Fatal(err)
	}

	// Generate non-centered layout
	renderer = svg.NewRenderer(decision_tree.DefaultConfig())
	renderer.SetCenterParent(false)
	svgContent = renderer.RenderTree(tree)
	err = os.WriteFile("tree_left.svg", []byte(svgContent), 0644)
	if err != nil {
		log.Fatal(err)
	}

	// Also save the tree definition as JSON for reference
	treeJSON, err := json.MarshalIndent(tree, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}
	err = os.WriteFile("tree.json", treeJSON, 0644)
	if err != nil {
		fmt.Printf("Error saving JSON: %v\n", err)
		return
	}

	fmt.Println("Generated tree.svg and tree.json")

	server := svg.NewServer(svg.NewRenderer(decision_tree.DefaultConfig()))
	err = server.Serve(tree)
	if err != nil {
		fmt.Printf("Error serving SVG: %v\n", err)
		return
	}
}
