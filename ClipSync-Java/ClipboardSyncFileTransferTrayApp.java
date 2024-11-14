import java.awt.*;
import java.awt.datatransfer.*;
import java.io.*;
import java.net.*;
import java.nio.file.*;
import java.util.List;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

public class ClipboardSyncFileTransferTrayApp {

    private static final int PORT = 12345;
    private static final String MULTICAST_GROUP = "230.0.0.1";
    private static String lastClipboardData = "";
    private static File lastClipboardFile = null;
    private static boolean isRunning = false;
    private static TrayIcon trayIcon;
    private static ExecutorService pool;
    private static MulticastSocket socket;
    private static InetAddress group;

    public static void main(String[] args) {
        try {
            if (!SystemTray.isSupported()) {
                System.err.println("SystemTray is not supported on this platform.");
                System.exit(1);
            }
            setupSystemTray();
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    private static void setupSystemTray() {
        SystemTray tray = SystemTray.getSystemTray();
        Image image = Toolkit.getDefaultToolkit().getImage("icon.gif");
        trayIcon = new TrayIcon(image, "Clipboard Sync");

        trayIcon.setImageAutoSize(true);

        PopupMenu popupMenu = new PopupMenu();

        MenuItem startItem = new MenuItem("Start Sync");
        MenuItem stopItem = new MenuItem("Stop Sync");
        MenuItem exitItem = new MenuItem("Exit");

        startItem.addActionListener(e -> startClipboardSync());

        stopItem.addActionListener(e -> stopClipboardSync());

        exitItem.addActionListener(e -> System.exit(0));

        popupMenu.add(startItem);
        popupMenu.add(stopItem);
        popupMenu.addSeparator();
        popupMenu.add(exitItem);

        trayIcon.setPopupMenu(popupMenu);

        try {
            tray.add(trayIcon);
        } catch (AWTException e) {
            System.err.println("TrayIcon could not be added.");
            e.printStackTrace();
        }
    }

    private static void startClipboardSync() {
        if (isRunning) {
            showTrayMessage("Clipboard sync is already running.");
            return;
        }
    
        isRunning = true;
        showTrayMessage("Clipboard Sync Started");
    
        pool = Executors.newFixedThreadPool(2);
    
        try {
            socket = new MulticastSocket(PORT);
            group = InetAddress.getByName(MULTICAST_GROUP);
            socket.joinGroup(group);
    
            String deviceName = InetAddress.getLocalHost().getHostName();
            String deviceId = "ID:" + deviceName;
    
            String joinMessage = deviceId + " JOIN:" + deviceName;
            sendMessage(joinMessage);
    
            Clipboard clipboard = Toolkit.getDefaultToolkit().getSystemClipboard();
    

            pool.execute(() -> {
                while (isRunning) {
                    try {
                        Transferable transferable = clipboard.getContents(null);
    
                        if (transferable != null && transferable.isDataFlavorSupported(DataFlavor.javaFileListFlavor)) {
                            @SuppressWarnings("unchecked")
                            List<File> fileList = (List<File>) transferable.getTransferData(DataFlavor.javaFileListFlavor);
                            if (!fileList.isEmpty()) {
                                File file = fileList.get(0);
                                if (!file.equals(lastClipboardFile)) {
                                    lastClipboardFile = file;
                                    String message = deviceId + " FILE:" + file.getName();
                                    sendFile(file, message);
                                }
                            }
                        } else if (transferable != null && transferable.isDataFlavorSupported(DataFlavor.stringFlavor)) {
                            String clipboardData = (String) transferable.getTransferData(DataFlavor.stringFlavor);
                            File file = new File(clipboardData);
                            if (file.exists() && file.isFile()) {
                                if (!file.equals(lastClipboardFile)) {
                                    lastClipboardFile = file;
                                    String message = deviceId + " FILE:" + file.getName();
                                    sendFile(file, message);
                                }
                            } else if (!clipboardData.equals(lastClipboardData)) {
                                lastClipboardData = clipboardData;
                                String message = deviceId + " TEXT:" + clipboardData;
                                sendMessage(message);
                            }
                        }
    
                        Thread.sleep(1000);
    
                    } catch (UnsupportedFlavorException e) {
                        showTrayMessage("Unsupported clipboard content: " + e.getMessage());
                    } catch (Exception e) {
                        e.printStackTrace();
                    }
                }
            });
    
            pool.execute(() -> {
                while (isRunning) {
                    try {
                        byte[] buffer = new byte[1024 * 64];
                        DatagramPacket packet = new DatagramPacket(buffer, buffer.length);
                        socket.receive(packet);
            
                        String receivedMessage = new String(packet.getData(), 0, packet.getLength());
            
                        String[] parts = receivedMessage.split(" ", 2);
                        String senderId = parts[0]; // Get sender ID (e.g., "ID:device-name")
                        String content = parts.length > 1 ? parts[1] : "";  // Get the actual message content
            
                        if (senderId.equals(deviceId)) {
                            continue;
                        }
            
                        if (content.startsWith("TEXT:")) {
                            // Handle received text clipboard data
                            String clipboardData = content.substring(5);
                            if (!clipboardData.equals(lastClipboardData)) {
                                lastClipboardData = clipboardData;
                                StringSelection selection = new StringSelection(clipboardData);
                                Toolkit.getDefaultToolkit().getSystemClipboard().setContents(selection, null);
                            }
                        } else if (content.startsWith("FILE:")) {
                            // Handle received file data
                            String[] fileInfo = content.split(":", 3); // We expect 3 parts: FILE, file name, file data
            
                            if (fileInfo.length < 3) {
                                showTrayMessage("Malformed file message received. Ignoring.");
                                continue;
                            }
            
                            String fileName = fileInfo[1];
                            byte[] fileData = fileInfo[2].getBytes(); // This part needs careful validation for data integrity
            
                            File downloadsFolder = getDownloadsFolder();
                            File receivedFile = new File(downloadsFolder, fileName);
                            Files.write(receivedFile.toPath(), fileData);
            
                            // showTrayMessage("File received: " + fileName);
                        } else if (content.startsWith("JOIN:")) {
                            // Handle JOIN message
                            String connectedDeviceName = content.substring(5);
                            showTrayMessage(connectedDeviceName + " is Connected to the ClipSync Group");
                        }
                    } catch (Exception e) {
                        e.printStackTrace();
                    }
                }
            });            
    
        } catch (Exception e) {
            showTrayMessage("Error starting clipboard sync: " + e.getMessage());
            stopClipboardSync();
        }
    }
    
    private static void sendFile(File file, String message) {
        try {
            byte[] fileData = Files.readAllBytes(file.toPath());
            message = message + ":" + new String(fileData);
            sendMessage(message);
    
            // showTrayMessage("File sent: " + file.getName());
    
        } catch (Exception e) {
            showTrayMessage("Error sending file: " + e.getMessage());
        }
    }

    private static void stopClipboardSync() {
        if (!isRunning) {
            showTrayMessage("Clipboard sync is not running.");
            return;
        }
        isRunning = false;
        showTrayMessage("Clipboard Sync Stopped");

        if (pool != null) {
            pool.shutdownNow();
        }
        if (socket != null && !socket.isClosed()) {
            try {
                socket.leaveGroup(group);
                socket.close();
            } catch (IOException e) {
                e.printStackTrace();
            }
        }
    }

    private static void sendMessage(String message) {
        try {
            DatagramPacket packet = new DatagramPacket(message.getBytes(), message.length(), group, PORT);
            socket.send(packet);
            System.out.println("Sent:" + message);
        } catch (Exception e) {
            e.printStackTrace();
        }
    }

    private static File getDownloadsFolder() {
        String userHome = System.getProperty("user.home");
        File downloadsFolder = new File(userHome, "Downloads");
        if (!downloadsFolder.exists()) {
            downloadsFolder.mkdir();
        }
        return downloadsFolder;
    }

    private static void showTrayMessage(String message) {
        if (trayIcon != null) {
            trayIcon.displayMessage("Clipboard Sync", message, TrayIcon.MessageType.INFO);
        }
    }
}