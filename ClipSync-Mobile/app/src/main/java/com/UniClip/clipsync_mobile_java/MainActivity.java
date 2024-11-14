package com.UniClip.clipsync_mobile_java;

import android.content.Intent;
import android.os.Bundle;
import android.widget.TextView;

import androidx.appcompat.app.AppCompatActivity;

public class MainActivity extends AppCompatActivity {

    private TextView textViewReceivedMessages;
    public static TextView staticTextView;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);



        textViewReceivedMessages = findViewById(R.id.textViewReceivedMessages);
        staticTextView = textViewReceivedMessages; // Static reference to update from the service

        // Start the background service to receive multicast messages
        Intent serviceIntent = new Intent(this, MulticastReceiverService.class);
        startService(serviceIntent);
    }
}