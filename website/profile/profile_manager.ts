import { map } from "rxjs";
import { Element, ElementConfig } from "../element";
import { SchemaManager } from "../schema_manager";
import { NewProfilePopup } from "./new_profile";
import { DropdownMenu } from "../components/dropdown";
import { RenameProfilePopup } from "./rename_profile";
import { OverwriteProfilePopup } from "./overwrite_profile";
import { DeleteProfilePopup } from "./delete_profile";
import { RequestManager } from "../requests";

function ProfileElement(
    profileName: string,
    schemaManager: SchemaManager,
    requestManager: RequestManager
): ElementConfig {
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
                            requestManager.applyProfile(
                                profileName,
                                () => schemaManager.refreshSchema("Applied a profile"),
                                () => alert("unable to load profile")
                            )
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
                                    const popoup = new RenameProfilePopup(profileName, schemaManager, requestManager);
                                    popoup.show();
                                }
                            },
                            {
                                text: "Overwrite",
                                onclick: () => {
                                    const popoup = new OverwriteProfilePopup(profileName, schemaManager, requestManager);
                                    popoup.show();
                                }
                            },
                            {
                                text: "Delete",
                                onclick: () => {
                                    const popoup = new DeleteProfilePopup(profileName, schemaManager, requestManager);
                                    popoup.show();
                                }
                            },
                        ]
                    }),

                ]
            },
        ]
    };
}

export class ProfileManager {
    constructor(
        parent: HTMLElement,
        schemaManager: SchemaManager,
        requestManager: RequestManager
    ) {
        const newProfileButton = parent.querySelector("#new-profile");

        newProfileButton.addEventListener('click', (event) => {
            const popup = new NewProfilePopup(schemaManager, requestManager);
            popup.show();
        });

        const variableListView = parent.querySelector("#profile-list")
        variableListView.append(Element({
            children$: schemaManager.
                schema$.
                pipe(map((graph): Array<ElementConfig> => {
                    return graph.
                        profiles?.
                        map((p) => ProfileElement(p, schemaManager, requestManager))
                })),
        }));
    }
}