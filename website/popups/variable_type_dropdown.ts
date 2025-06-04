import { BehaviorSubject } from "rxjs";
import { ElementConfig } from "../element";
import { VariableType } from "../variable_type";

export function VariableTypeDropdown(change$: BehaviorSubject<string>): ElementConfig {
    return {
        tag: "select",
        change$: change$,
        children: [
            { tag: "option", value: VariableType.Float, text: "Float" },
            { tag: "option", value: VariableType.Float2, text: "Float2" },
            { tag: "option", value: VariableType.Float3, text: "Float3" },
            { tag: "option", value: VariableType.Int, text: "Int" },
            { tag: "option", value: VariableType.Int2, text: "Int2" },
            { tag: "option", value: VariableType.Int3, text: "Int3" },
            { tag: "option", value: VariableType.String, text: "String" },
            { tag: "option", value: VariableType.Bool, text: "Bool" },
            { tag: "option", value: VariableType.AABB, text: "AABB" },
            { tag: "option", value: VariableType.Color, text: "Color" },
        ]
    };
}