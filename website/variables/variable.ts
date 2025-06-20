import { DropdownMenu } from "../components/dropdown";
import { ElementConfig } from "../element";
import { NodeManager } from "../node_manager";
import { DeleteVariablePopup } from "../popups/delete_variable";
import { EditVariablePopup } from "../popups/edit_variable";
import { Variable } from "../schema";
import { SchemaManager } from "../schema_manager";
import { ElementInstance } from "./element_instance";

export abstract class VariableElement extends ElementInstance<Variable> {

    constructor(
        protected key: string,
        protected variable: Variable,
        private schemaManager: SchemaManager,
        private nodeManager: NodeManager,
    ) {
        super();
    }

    abstract buildVariable(): ElementConfig;

    build(): ElementConfig {
        let input: ElementConfig = this.buildVariable();

        return {
            style: {
                marginTop: "16px",
                display: "flex",
                flexDirection: "row"
            },
            children: [
                {
                    style: {
                        display: "flex",
                        flexDirection: "column",
                        flexGrow: "1"
                    },
                    children: [
                        {
                            style: {
                                display: "flex",
                                flexDirection: "row"
                            },
                            children: [
                                {
                                    text: this.key,
                                    classList: ["variable-name"],
                                },
                                DropdownMenu({
                                    buttonContent: {
                                        tag: "i",
                                        classList: ["fa-solid", "fa-ellipsis-vertical"]
                                    },
                                    buttonClasses: ["icon-button"],
                                    content: [
                                        {
                                            text: "Edit",
                                            onclick: () => {
                                                const popoup = new EditVariablePopup(this.key, this.variable);
                                                popoup.show();
                                            }
                                        },
                                        {
                                            text: "Delete",
                                            onclick: () => {
                                                const deletePopoup = new DeleteVariablePopup(this.schemaManager, this.nodeManager, this.key, this.variable);
                                                deletePopoup.show();
                                            }
                                        },
                                    ]
                                }),

                            ]
                        },
                        {
                            text: this.variable.description,
                            classList: ["variable-description"],

                        },
                        input
                    ]
                }
            ]
        };
    }
}