import { map } from "rxjs";
import { Element, ElementConfig } from "../element";
import { SchemaManager } from "../schema_manager";
import { NewProfilePopup } from "./new_profile";
import { DropdownMenu } from "../components/dropdown";
import { RenameProfilePopup } from "./rename_profile";
import { OverwriteProfilePopup } from "./overwrite_profile";
import { DeleteProfilePopup } from "./delete_profile";

function ProfileElement(profileName: string, schemaManager: SchemaManager): ElementConfig {
    return {
        style: {
            // marginBottom: "16px",
        },
        children: [
            {
                style: {
                    display: "flex",
                    flexDirection: "row"
                },
                children: [
                    {
                        text: profileName,
                        classList: ["variable-name", "profile-item"],
                        onclick: () => {
                            fetch("./profile/apply", {
                                method: "POST",
                                body: JSON.stringify({ name: profileName, })
                            }).then((resp) => {
                                if (!resp.ok) {
                                    alert("unable to load profile");
                                } else {
                                    schemaManager.refreshSchema();
                                }
                            });
                        }
                    },
                    DropdownMenu({
                        buttonContent: {
                            tag: "i",
                            classList: ["fa-solid", "fa-ellipsis-vertical"]
                        },
                        buttonClasses: ["icon-button"],
                        content: [
                            {
                                text: "Rename",
                                onclick: () => {
                                    const popoup = new RenameProfilePopup(profileName, schemaManager);
                                    popoup.show();
                                }
                            },
                            {
                                text: "Overwrite",
                                onclick: () => {
                                    const popoup = new OverwriteProfilePopup(profileName, schemaManager);
                                    popoup.show();
                                    // const popoup = new EditVariablePopup(this.key, this.variable);
                                    // popoup.show();
                                }
                            },
                            {
                                text: "Delete",
                                onclick: () => {
                                    const popoup = new DeleteProfilePopup(profileName, schemaManager);
                                    popoup.show();
                                    // const deletePopoup = new DeleteVariablePopup(this.schemaManager, this.nodeManager, this.key, this.variable);
                                    // deletePopoup.show();
                                }
                            },
                        ]
                    }),

                ]
            },
            // {
            //     tag: "button",
            //     text: "Apply",
            //     onclick: () => {
            //         fetch("./profile-apply", {
            //             method: "POST",
            //             body: JSON.stringify({ name: profileName, })
            //         }).then((resp) => {
            //             if (!resp.ok) {
            //                 alert("unable to load profile");
            //             } else {
            //                 schemaManager.refreshSchema();
            //             }
            //         });
            //     }
            // }
        ]
    };
}

export class ProfileManager {
    constructor(parent: HTMLElement, schemaManager: SchemaManager) {
        const newProfileButton = parent.querySelector("#new-profile");

        newProfileButton.addEventListener('click', (event) => {
            const popup = new NewProfilePopup(schemaManager);
            popup.show();
        });

        const variableListView = parent.querySelector("#profile-list")
        variableListView.append(Element({
            children$: schemaManager.
                schema$.
                pipe(map((graph): Array<ElementConfig> => {
                    return graph.
                        profiles?.
                        map((p) => ProfileElement(p, schemaManager))
                })),
        }));
    }
}