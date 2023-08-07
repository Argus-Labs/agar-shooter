using System;
using UnityEngine;

public class CircularArray<T>
{
    private T[] array;  
    private int head;  
    private int tail;  
    private int count;  

    public int Count => count;  

    public CircularArray(int capacity=500)
    {
        array = new T[capacity];
        head = 0;
        tail = 0;
        count = 0;
    }

    public void Enqueue(T item)
    {
        if (count == array.Length)
        {
            // 扩容数组
            ResizeArray();
        }

        array[tail] = item;
        tail = (tail + 1) % array.Length;
        count++;
    }

    public bool IsFull()
    {
        // because enqueue will resize the array when it is full
        return false;
    }

    public T Dequeue()
    {
        if (count == 0)
        {
            throw new InvalidOperationException("Queue is empty.");
        }

        T item = array[head];
        head = (head + 1) % array.Length;
        count--;
        return item;
    }

    public T Peek()
    {
        if (count == 0)
        {
            throw new InvalidOperationException("Queue is empty.");
        }

        return array[head];
    }

    public bool IsEmpty()
    {
        return count == 0;
    }

    public void Clear()
    {
        head = 0;
        tail = 0;
        count = 0;
    }

    public T this[int index]
    {
        get
        {
            if (index < 0 || index >= count)
            {
                throw new IndexOutOfRangeException("Index is out of range.");
            }

            int actualIndex = (head + index) % array.Length;
            return array[actualIndex];
        }
        set
        {
             if (index < 0 || index >= count)
             {
                 throw new IndexOutOfRangeException("Index is out of range.");
             }

             int actualIndex = (head + index) % array.Length;
             array[actualIndex] = value;
        }
    }

    private void ResizeArray()
    {
        Debug.LogError($"{array.Length} is not enough, now extend to{array.Length*2} ");
        T[] newArray = new T[array.Length * 2];
        if (head < tail)
        {
            Array.Copy(array, head, newArray, 0, count);
        }
        else
        {
            Array.Copy(array, head, newArray, 0, array.Length - head);
            Array.Copy(array, 0, newArray, array.Length - head, tail);
        }

        head = 0;
        tail = count;
        array = newArray;
    }
}