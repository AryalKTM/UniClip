import java.awt.*;
import java.awt.datatransfer.*;
import java.io.IOException;

public class ClipboardMonitor {
    private String previousContent = "";

    public interface ClipboardListener {
        void onClipboardChange(String content);
    }

    public void startMonitoring(ClipboardListener listener) {
        Clipboard clipboard = Toolkit.getDefaultToolkit().getSystemClipboard();

        new Thread(() -> {
            while (true) {
                try {
                    String currentContent = getClipboardContent(clipboard);
                    if (!currentContent.equals(previousContent)) {
                        previousContent = currentContent;
                        listener.onClipboardChange(currentContent);
                    }

                    Thread.sleep(1000);
                } catch (Exception e) {
                    e.printStackTrace();
                }
            }
        }).start();
    }

    private String getClipboardContent(Clipboard clipboard) {
        try {
            Transferable content = clipboard.getContents(null);
            if (content != null && content.isDataFlavorSupported(DataFlavor.stringFlavor)) {
                return (String) content.getTransferData(DataFlavor.stringFlavor);
            }
        } catch (UnsupportedFlavorException | IOException e) {
            e.printStackTrace();
        }
        return "";
    }

    public static void setClipboardContent(String content) {
        StringSelection stringSelection = new StringSelection(content);
        Clipboard clipboard = Toolkit.getDefaultToolkit().getSystemClipboard();
        clipboard.setContents(stringSelection, null);
    }
}