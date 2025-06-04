import { Element, ElementConfig } from './element';
import { NewVariablePopup } from "./popups/new_variable";
import { GraphInstance, Variable } from "./schema";
import { SchemaManager } from "./schema_manager";
import { VariableType } from './variable_type';
import { BehaviorSubject, combineLatestWith, flatMap, map, mergeMap, Observable, skip, Subject } from "rxjs";
import { NodeManager } from './node_manager';
import { DeleteVariablePopup } from './popups/delete_variable';
import { EditVariablePopup } from './popups/edit_variable';

const inputStyle: Partial<CSSStyleDeclaration> = {
    flexShrink: "1",
    flexGrow: "1",
    minWidth: "0",
    flexBasis: "0"
}

const inputContainerStyle: Partial<CSSStyleDeclaration> = {
    display: "flex",
    flexDirection: "column",
    flexShrink: "1",
    paddingLeft: "8px",
    paddingRight: "8px",
}

export function post$(url: string, body: BodyInit): Observable<Response> {
    const out = new Subject<Response>();
    fetch(url, {
        method: "POST",
        body: body
    }).then((resp) => {
        out.next(resp);
    });
    return out;
}


export class VariableManager {

    variableListView: Element;

    nodeManager: NodeManager;

    schemaManager: SchemaManager;

    constructor(parent: HTMLElement, schemaManager: SchemaManager, nodeManager: NodeManager) {
        this.nodeManager = nodeManager;
        this.schemaManager = schemaManager;
        const newVariableButton = parent.querySelector("#new-variable")
        // const newFolderButton = parent.querySelector("#new-folder")
        this.variableListView = parent.querySelector("#variable-list")

        newVariableButton.addEventListener('click', (event) => {
            const popup = new NewVariablePopup(schemaManager, this.nodeManager);
            popup.show();
        });

        schemaManager.subscribe(this.newSchemaInstance.bind(this));
    }

    newSchemaInstance(graphInstance: GraphInstance): void {
        let arr = new Array<Element>();
        for (const variableKey in graphInstance.variables.variables) {
            const variable = graphInstance.variables.variables[variableKey];
            arr.push(this.newVariable(variableKey, variable));
        }
        this.variableListView.replaceChildren(...arr);
    }

    setVariableValue(variable: string, value: any): Observable<Response> {
        return post$("./variable/value/" + variable, JSON.stringify(value))
    }

    newBasicVariable<T>(key: string, variable: Variable, mapper: (s: string) => T): ElementConfig {
        const variableTopic = new Subject<string>();

        variableTopic.pipe(
            map(mapper),
            mergeMap((val) => this.setVariableValue(key, val))
        ).subscribe((resp: Response) => {
            console.log(resp);
        })

        return {
            tag: "input",
            change$: variableTopic,
            value: `${variable.value}`,
            size: 1,
            style: {
                minWidth: "0",
                flexShrink: "1",
                flexGrow: "1"
            }
        };
    }

    newVector2Variable<T>(key: string, variable: Variable, mapper: (s: string) => T, step: string): ElementConfig {
        const x = new BehaviorSubject<string>(`${variable.value.x}`);
        const y = new BehaviorSubject<string>(`${variable.value.y}`);

        x.pipe(
            map(mapper),
            combineLatestWith(y.pipe(map(mapper))),
            skip(1),
            mergeMap((val) => this.setVariableValue(key, { x: val[0], y: val[1] }))
        ).subscribe((resp: Response) => {
            console.log(resp);
        })

        return {
            style: inputContainerStyle,
            children: [
                { tag: "input", change$: x, type: "number", style: inputStyle, value: `${variable.value.x}`, step: step },
                { tag: "input", change$: y, type: "number", style: inputStyle, value: `${variable.value.y}`, step: step },
            ]
        };
    }

    newVector3Variable<T>(key: string, variable: Variable, mapper: (s: string) => T, step: string): ElementConfig {
        const x = new BehaviorSubject<string>(`${variable.value.x}`);
        const y = new BehaviorSubject<string>(`${variable.value.y}`);
        const z = new BehaviorSubject<string>(`${variable.value.z}`);

        x.pipe(
            map(mapper),
            combineLatestWith(y.pipe(map(mapper)), z.pipe(map(mapper))),
            skip(1), // Ignore the first change, as it's just the initial value
            mergeMap((val) => this.setVariableValue(key, { x: val[0], y: val[1], z: val[2] }))
        ).subscribe((resp: Response) => {
            console.log(resp);
        });

        return {
            style: inputContainerStyle,
            children: [
                { tag: "input", change$: x, type: "number", size: 1, style: inputStyle, value: `${variable.value.x}`, step: step },
                { tag: "input", change$: y, type: "number", size: 1, style: inputStyle, value: `${variable.value.y}`, step: step },
                { tag: "input", change$: z, type: "number", size: 1, style: inputStyle, value: `${variable.value.z}`, step: step },
            ]
        };
    }

    newAABBVariable(key: string, variable: Variable): ElementConfig {
        const centerx = new BehaviorSubject<string>(`${variable.value.center.x}`);
        const centery = new BehaviorSubject<string>(`${variable.value.center.y}`);
        const centerz = new BehaviorSubject<string>(`${variable.value.center.z}`);

        const extentsx = new BehaviorSubject<string>(`${variable.value.extents.x}`);
        const extentsy = new BehaviorSubject<string>(`${variable.value.extents.y}`);
        const extentsz = new BehaviorSubject<string>(`${variable.value.extents.z}`);
        centerx.pipe(
            map(parseFloat),
            combineLatestWith(
                centery.pipe(map(parseFloat)),
                centerz.pipe(map(parseFloat)),
                extentsx.pipe(map(parseFloat)),
                extentsy.pipe(map(parseFloat)),
                extentsz.pipe(map(parseFloat))
            ),
            skip(1), // Ignore the first change, as it's just the initial value
            mergeMap((val) => this.setVariableValue(key, {
                center: {
                    x: val[0], y: val[1], z: val[2]
                },
                extents: {
                    x: val[3], y: val[4], z: val[5]
                },
            }))
        ).subscribe((resp: Response) => {
            console.log(resp);
        })

        return {
            style: inputContainerStyle,
            children: [
                { text: "center" },
                { tag: "input", change$: centerx, type: "number", style: inputStyle, value: `${variable.value.center.x}` },
                { tag: "input", change$: centery, type: "number", style: inputStyle, value: `${variable.value.center.y}` },
                { tag: "input", change$: centerz, type: "number", style: inputStyle, value: `${variable.value.center.z}` },

                { text: "extents" },
                { tag: "input", change$: extentsx, type: "number", style: inputStyle, value: `${variable.value.extents.x}` },
                { tag: "input", change$: extentsy, type: "number", style: inputStyle, value: `${variable.value.extents.y}` },
                { tag: "input", change$: extentsz, type: "number", style: inputStyle, value: `${variable.value.extents.z}` },
            ]
        };
    }

    newVariable(key: string, variable: Variable): Element {
        console.log(variable)

        const intMap = (s: string) => parseInt(s)

        let input: ElementConfig;
        switch (variable.type) {
            case VariableType.Float:
                input = this.newBasicVariable(key, variable, parseFloat);
                input.type = "number";
                break;

            case VariableType.Float2:
                input = this.newVector2Variable(key, variable, parseFloat, "");
                break;

            case VariableType.Float3:
                input = this.newVector3Variable(key, variable, parseFloat, "");
                break;

            case VariableType.Int2:
                input = this.newVector2Variable(key, variable, intMap, "1");
                break;

            case VariableType.Int3:
                input = this.newVector3Variable(key, variable, intMap, "1");
                break;

            case VariableType.Int:
                input = this.newBasicVariable(key, variable, intMap);
                input.type = "number";
                input.step = "1";
                break;

            case VariableType.String:
                input = this.newBasicVariable(key, variable, (s) => s);
                break;

            case VariableType.Color:
                input = this.newBasicVariable(key, variable, (s) => s);
                input.type = "color";
                break;

            case VariableType.Bool:
                input = this.newBasicVariable(key, variable, (s) => s === "true");
                input.type = "checkbox";
                break;

            case VariableType.AABB:
                input = this.newAABBVariable(key, variable);
                break;

            default:
                throw new Error("unimplemented variable type: " + variable.type);
        }

        return Element({
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
                        justifyContent: "center",
                        marginRight: "8px"
                    },
                    children: [
                        {
                            tag: "button",
                            text: "Edit",
                            onclick: () => {
                                const popoup = new EditVariablePopup(this.schemaManager, this.nodeManager, key, variable);
                                popoup.show();
                            }
                        },
                        {
                            tag: "button",
                            text: "Delete",
                            onclick: () => {
                                const deletePopoup = new DeleteVariablePopup(this.schemaManager, this.nodeManager, key, variable);
                                deletePopoup.show();
                            }
                        }
                    ]
                },
                {
                    style: {
                        display: "flex",
                        flexDirection: "column",
                        flexGrow: "1"
                    },
                    children: [
                        {
                            text: variable.name,
                            style: {
                                textDecoration: "underline"
                            }
                        },
                        {
                            text: variable.description,
                            style: {
                                lineHeight: "normal",
                                marginBottom: "8px",
                            }
                        },
                        input
                    ]
                }
            ]
        })
    }
}