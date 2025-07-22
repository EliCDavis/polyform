using System;
using System.Collections.Generic;
using UnityEngine;
using UnityEngine.UIElements;

namespace EliCDavis.Polyform.Elements
{
    public class Vector3ArrayField : VisualElement
    {
        private List<Vector3> values = new List<Vector3>();
        private readonly VisualElement listContainer;

        public event Action<List<Vector3>> OnValueChanged;

        public Vector3ArrayField(string label, IEnumerable<Vector3> initialValues = null)
        {
            style.flexDirection = FlexDirection.Column;
            style.marginBottom = 10;

            var header = new Label(string.IsNullOrWhiteSpace(label) ? "Vector3 Array" : label);
            Add(header);

            listContainer = new VisualElement
            {
                style = { flexDirection = FlexDirection.Column, marginLeft = 10 }
            };
            Add(listContainer);

            var addButton = new Button(AddItem)
            {
                text = "+ Add"
            };
            addButton.style.marginTop = 4;
            Add(addButton);
            // Debug.Log(initialValues);
            if (initialValues != null)
                values.AddRange(initialValues);

            RefreshList();
        }

        private void RefreshList()
        {
            listContainer.Clear();

            for (int i = 0; i < values.Count; i++)
            {
                int index = i;
                var item = new VisualElement { style = { flexDirection = FlexDirection.Row } };

                var vecField = new Vector3Field { value = values[i] };
                vecField.style.flexGrow = 1;

                vecField.RegisterValueChangedCallback(evt =>
                {
                    values[index] = evt.newValue;
                    OnValueChanged?.Invoke(new List<Vector3>(values));
                });

                var removeBtn = new Button(() =>
                {
                    values.RemoveAt(index);
                    RefreshList();
                    OnValueChanged?.Invoke(new List<Vector3>(values));
                })
                {
                    text = "X"
                };
                removeBtn.style.marginLeft = 4;

                item.Add(vecField);
                item.Add(removeBtn);
                listContainer.Add(item);
            }
        }

        private void AddItem()
        {
            values.Add(Vector3.zero);
            RefreshList();
            OnValueChanged?.Invoke(new List<Vector3>(values));
        }

        public void SetValues(IEnumerable<Vector3> newValues)
        {
            values = new List<Vector3>(newValues);
            RefreshList();
            OnValueChanged?.Invoke(new List<Vector3>(values));
        }

        public List<Vector3> GetValues()
        {
            return new List<Vector3>(values);
        }
    }
}