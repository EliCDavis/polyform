using System;
using UnityEngine;

namespace EliCDavis.Polyform.Utils
{
    [Serializable]
    internal class WeightedListConfig<T>
    {
        [SerializeField] private WeightedListConfigItem<T>[] items;

        public WeightedList<T> List()
        {
            var result = new WeightedListItem<T>[items.Length];
            for (var i = 0; i < items.Length; i++)
            {
                result[i] = items[i].Item();
            }

            return new WeightedList<T>(result);
        }
    }

    [Serializable]
    public class WeightedListConfigItem<T>
    {
        [SerializeField] private T item;

        [SerializeField] private int weight;

        public WeightedListItem<T> Item()
        {
            return new WeightedListItem<T>(item, weight);
        }
    }
}