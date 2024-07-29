package com.example.myapplication;

import android.content.ClipboardManager;
import android.content.Context;
import android.os.AsyncTask;
import android.os.Bundle;
import android.text.method.ScrollingMovementMethod;
import android.view.View;
import android.widget.Button;
import android.widget.EditText;
import android.widget.TextView;
import androidx.appcompat.app.AppCompatActivity;
import java.io.BufferedReader;
import java.io.BufferedWriter;
import java.io.InputStreamReader;
import java.io.OutputStreamWriter;
import java.io.PrintWriter;
import java.net.InetAddress;
import java.net.ServerSocket;
import java.net.Socket;

public class MainActivity extends AppCompatActivity {
    private EditText editTextAddress;
    private Button buttonStartServer, buttonConnect;
    private TextView textViewLog;
    private ServerSocket serverSocket;
    private Socket clientSocket;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        editTextAddress = findViewById(R.id.editTextAddress);
        buttonStartServer = findViewById(R.id.buttonStartServer);
        buttonConnect = findViewById(R.id.buttonConnect);
        textViewLog = findViewById(R.id.textViewLog);
        textViewLog.setMovementMethod(new ScrollingMovementMethod());

        buttonStartServer.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                new Thread(new Runnable() {
                    @Override
                    public void run() {
                        startServer();
                    }
                }).start();
            }
        });

        buttonConnect.setOnClickListener(new View.OnClickListener() {
            @Override
            public void onClick(View v) {
                String address = editTextAddress.getText().toString();
                new Thread(new Runnable() {
                    @Override
                    public void run() {
                        connectToServer(address);
                    }
                }).start();
            }
        });
    }

    private void logMessage(final String message) {
        runOnUiThread(new Runnable() {
            @Override
            public void run() {
                textViewLog.append(message + "\n");
            }
        });
    }

    private void startServer() {
        try {
            serverSocket = new ServerSocket(6000);
            logMessage("Server started. Waiting for clients...");
            while (true) {
                clientSocket = serverSocket.accept();
                logMessage("Client connected: " + clientSocket.getInetAddress().getHostAddress());

                runOnUiThread(new Runnable() {
                    @Override
                    public void run() {
                        new ClientHandlerTask(clientSocket).executeOnExecutor(AsyncTask.THREAD_POOL_EXECUTOR);
                    }
                });
            }
        } catch (Exception e) {
            logMessage("Error: " + e.getMessage());
        }
    }

    private void connectToServer(String address) {
        try {
            String[] parts = address.split(":");
            String ipAddress = parts[0];
            int port = Integer.parseInt(parts[1]);

            InetAddress serverAddr = InetAddress.getByName(ipAddress);
            clientSocket = new Socket(serverAddr, port);
            logMessage("Connected to server at " + address);
            new ClientHandlerTask(clientSocket).executeOnExecutor(AsyncTask.THREAD_POOL_EXECUTOR);
        } catch (Exception e) {
            logMessage("Error: " + e.getMessage());
        }
    }

    private class ClientHandlerTask extends AsyncTask<Void, String, Void> {
        private Socket socket;
        private BufferedReader in;
        private PrintWriter out;

        ClientHandlerTask(Socket socket) {
            this.socket = socket;
        }

        @Override
        protected Void doInBackground(Void... voids) {
            try {
                in = new BufferedReader(new InputStreamReader(socket.getInputStream()));
                out = new PrintWriter(new BufferedWriter(new OutputStreamWriter(socket.getOutputStream())), true);

                String line;
                while ((line = in.readLine()) != null) {
                    publishProgress(line);
                }
            } catch (Exception e) {
                publishProgress("Error: " + e.getMessage());
            }
            return null;
        }

        @Override
        protected void onProgressUpdate(String... values) {
            logMessage("Received: " + values[0]);

            // Set clipboard content on the Android device
            ClipboardManager clipboard = (ClipboardManager) getSystemService(Context.CLIPBOARD_SERVICE);
            clipboard.setPrimaryClip(android.content.ClipData.newPlainText("Copied Text", values[0]));
        }

        @Override
        protected void onPostExecute(Void aVoid) {
            try {
                socket.close();
            } catch (Exception e) {
                logMessage("Error: " + e.getMessage());
            }
        }
    }
}
