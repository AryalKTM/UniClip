package com.UniClip.clipsync_mobile_java;

import android.app.Service;
import android.content.Intent;
import android.net.wifi.WifiManager;
import android.os.IBinder;
import android.util.Log;

import java.io.IOException;
import java.net.DatagramPacket;
import java.net.InetAddress;
import java.net.MulticastSocket;

public class MulticastReceiverService extends Service {

    private static final String TAG = "MulticastReceiverService";
    private static final String MULTICAST_IP = "230.0.0.1";
    private static final int PORT = 12345;
    private boolean isReceiving = true;
    private MulticastSocket multicastSocket;
    private WifiManager.MulticastLock multicastLock; // Declare multicastLock here

    @Override
    public void onCreate() {
        super.onCreate();
        // Acquire the multicast lock
        WifiManager wifiManager = (WifiManager) getSystemService(WIFI_SERVICE);
        if (wifiManager != null) {
            multicastLock = wifiManager.createMulticastLock("MulticastLock");
            multicastLock.acquire();
        }
        startReceivingMessages();
    }

    private void startReceivingMessages() {
        new Thread(() -> {
            try {
                InetAddress group = InetAddress.getByName(MULTICAST_IP);
                multicastSocket = new MulticastSocket(PORT);
                multicastSocket.joinGroup(group);
                Log.d(TAG, "Joined multicast group: " + MULTICAST_IP + ":" + PORT);

                byte[] buffer = new byte[1024];
                while (isReceiving) {
                    DatagramPacket packet = new DatagramPacket(buffer, buffer.length);
                    multicastSocket.receive(packet);
                    final String receivedMessage = new String(packet.getData(), 0, packet.getLength());
                    Log.d(TAG, "Received message: " + receivedMessage);

                    // Handle received message (e.g., update UI if applicable)
                }
            } catch (IOException e) {
                Log.e(TAG, "Error receiving multicast message", e);
            }
        }).start();
    }

    @Override
    public void onDestroy() {
        super.onDestroy();
        isReceiving = false;
        // Release the multicast lock
        if (multicastLock != null && multicastLock.isHeld()) {
            multicastLock.release();
        }
        // Close multicast socket
        if (multicastSocket != null) {
            try {
                multicastSocket.leaveGroup(InetAddress.getByName(MULTICAST_IP));
                multicastSocket.close();
            } catch (IOException e) {
                Log.e(TAG, "Error closing multicast socket", e);
            }
        }
    }

    @Override
    public IBinder onBind(Intent intent) {
        return null; // Not used in this example
    }
}