using UnityEngine;

[CreateAssetMenu(fileName = "DefaultProfile", menuName = "ScriptableObjects/inputProfile", order = 1)]
public class InputProfile : ScriptableObject
{
    public KeyCode up = KeyCode.W;
    public KeyCode down = KeyCode.S;
    public KeyCode left = KeyCode.A;
    public KeyCode right = KeyCode.D;
}