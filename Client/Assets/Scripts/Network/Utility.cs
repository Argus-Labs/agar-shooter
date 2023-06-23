using UnityEngine;

public static class Utility
{
    public static string DeviceUniqueIdentifier
    {
        get
        {
            var deviceId = "";
 
 
#if UNITY_EDITOR
            deviceId = SystemInfo.deviceUniqueIdentifier + "-editor";

#elif UNITY_WEBGL
                if (!PlayerPrefs.HasKey("UniqueIdentifier"))
                    PlayerPrefs.SetString("UniqueIdentifier", Guid.NewGuid().ToString());
                deviceId = PlayerPrefs.GetString("UniqueIdentifier");
#else
                deviceId = SystemInfo.deviceUniqueIdentifier;
#endif
            return deviceId;
        }
    }
}