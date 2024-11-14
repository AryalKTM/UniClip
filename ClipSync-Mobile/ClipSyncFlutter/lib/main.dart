import 'dart:async';
import 'dart:io';
import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart'; // For Clipboard
import 'package:path_provider/path_provider.dart'; // For getting the Downloads directory
import 'package:receive_sharing_intent_plus/receive_sharing_intent_plus.dart'; // Updated import

void main() => runApp(const MulticastApp());

class MulticastApp extends StatefulWidget {
  const MulticastApp({super.key});

  @override
  _MulticastAppState createState() => _MulticastAppState();
}

class _MulticastAppState extends State<MulticastApp> {
  List<MessageData> messages = [];
  RawDatagramSocket? _socket;
  StreamSubscription? _intentDataStreamSubscription;

  @override
  void initState() {
    super.initState();
    _startListening();

    _intentDataStreamSubscription = ReceiveSharingIntentPlus.getTextStream().listen((String? sharedText) {
      if (sharedText != null) {
        String message = 'TEXT:$sharedText';
        _sendMessage(message);
        setState(() {
          messages.add(MessageData(id: 'Shared', content: sharedText));
        });
      }
    }, onError: (err) {
      print("Error in getTextStream: $err");
    });

    _intentDataStreamSubscription = ReceiveSharingIntentPlus.getMediaStream().listen((List<SharedMediaFile>? sharedFiles) {
      if (sharedFiles != null && sharedFiles.isNotEmpty) {
        for (var file in sharedFiles) {
          File fileObject = File(file.path);
          fileObject.readAsBytes().then((fileData) {
            String message = 'FILE:${file.path}:${String.fromCharCodes(fileData)}';
            _sendMessage(message);
            setState(() {
              messages.add(MessageData(id: 'Shared', content: 'File: ${file.path}'));
            });
          }).catchError((err) {
            print("Error reading file: $err");
          });
        }
      }
    }, onError: (err) {
      print("Error in getMediaStream: $err");
    });
  }

  @override
  void dispose() {
    _socket?.close();
    _intentDataStreamSubscription?.cancel();
    super.dispose();
  }

  void _startListening() async {
    const multicastAddress = '230.0.0.1';
    const port = 12345;

    _socket = await RawDatagramSocket.bind(InternetAddress.anyIPv4, port);
    _socket?.joinMulticast(InternetAddress(multicastAddress));

    _socket?.listen((RawSocketEvent event) {
      if (event == RawSocketEvent.read) {
        Datagram? dg = _socket?.receive();
        if (dg != null) {
          String message = String.fromCharCodes(dg.data);
          String id = _extractId(message);

          if (message.contains('TEXT:')) {
            String extractedText = _extractText(message);
            Clipboard.setData(ClipboardData(text: extractedText));
            setState(() {
              messages.add(MessageData(id: id, content: extractedText));
            });
          } else if (message.contains('FILE:')) {
            _saveFile(dg.data);
            setState(() {
              messages.add(MessageData(id: id, content: 'File received and saved.'));
            });
          }
        }
      }
    });
  }

  void _sendMessage(String message) {
    const multicastAddress = '230.0.0.1';
    const port = 12345;

    if (_socket != null) {
      try {
        List<int> data = message.codeUnits;
        _socket!.send(data, InternetAddress(multicastAddress), port);
        print('Message sent: $message');
      } catch (e) {
        print('Error sending message: $e');
      }
    } else {
      print('Socket is not initialized');
    }
  }

  String _extractId(String message) {
    try {
      int idIndex = message.indexOf('ID:') + 3;
      int endIndex = message.indexOf(' ', idIndex);
      if (idIndex > 3) {
        if (endIndex == -1) endIndex = message.length;
        return message.substring(idIndex, endIndex).trim();
      }
    } catch (e) {
      return 'Unknown';
    }
    return 'Unknown';
  }

  String _extractText(String message) {
    try {
      int textIndex = message.indexOf('TEXT:') + 5;
      if (textIndex > 5) {
        return message.substring(textIndex).trim();
      }
    } catch (e) {
      return 'Error extracting text';
    }
    return 'No text found';
  }

  Future<void> _saveFile(Uint8List data) async {
    try {
      String message = String.fromCharCodes(data);
      int fileNameStartIndex = message.indexOf('FILE:') + 5;
      int fileNameEndIndex = message.indexOf(':', fileNameStartIndex);

      if (fileNameEndIndex == -1) {
        setState(() {
          messages.add(MessageData(id: '', content: 'Error: Invalid file format.'));
        });
        return;
      }

      String fileName = message.substring(fileNameStartIndex, fileNameEndIndex).trim();
      Uint8List fileData = data.sublist(fileNameEndIndex + 1);

      Directory? downloadsDir = await getDownloadsDirectory();
      String filePath = '${downloadsDir?.path}/$fileName';
      File file = File(filePath);
      await file.writeAsBytes(fileData);

      setState(() {
        messages.add(MessageData(id: '', content: 'File saved to: $filePath'));
      });
    } catch (e) {
      setState(() {
        messages.add(MessageData(id: '', content: 'Error saving file: $e'));
      });
    }
  }

  void _copyToClipboard(String text) {
    Clipboard.setData(ClipboardData(text: text));
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Copied to clipboard')),
    );
  }

  void _deleteMessage(int index) {
    setState(() {
      messages.removeAt(index);
    });
  }

  String _getSnippet(String content) {
    return content.length > 50 ? '${content.substring(0, 50)}...' : content;
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        brightness: Brightness.dark,
        primaryColor: const Color(0xFF05245b),
        colorScheme: const ColorScheme(
          brightness: Brightness.dark,
          primary: Color(0xFF05245b),
          secondary: Color(0xFF04235b),
          surface: Color(0xFF050e1f),
          error: Colors.red,
          onPrimary: Colors.white,
          onSecondary: Colors.white,
          onSurface: Colors.white,
          onError: Colors.white,
        ),
        appBarTheme: const AppBarTheme(
          backgroundColor: Color(0xFF05245b),
          elevation: 0,
        ),
        cardColor: const Color(0xFF04235b),
        buttonTheme: ButtonThemeData(
          buttonColor: const Color(0xFF05245b),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(20),
          ),
        ),
        elevatedButtonTheme: ElevatedButtonThemeData(
          style: ElevatedButton.styleFrom(
            backgroundColor: const Color(0xFFe6af2c),
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(20),
            ),
          ),
        ),
        snackBarTheme: const SnackBarThemeData(
          backgroundColor: Color(0xFF05245b),
        ),
      ),
      home: Scaffold(
        appBar: AppBar(
          title: const Text('ClipSync', style: TextStyle(color: Color(0xFFe6af2c)),),
        ),
        body: Column(
          children: [
            Expanded(
              child: ListView.builder(
                itemCount: messages.length,
                itemBuilder: (context, index) {
                  return Container(
                    margin: const EdgeInsets.symmetric(vertical: 4.0), // Reduced vertical margin
                    child: ClipRRect(
                      borderRadius: BorderRadius.circular(20),
                      child: Card(
                        color: const Color(0xFF04235b),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(20),
                        ),
                        elevation: 0,  // Ensure no shadow
                        child: ExpansionTile(
                          title: Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              Expanded(
                                child: Text(
                                  _getSnippet(messages[index].content),
                                  style: const TextStyle(
                                    fontWeight: FontWeight.bold,
                                    color: Colors.white,
                                  ),
                                ),
                              ),
                              Row(
                                children: [
                                  IconButton(
                                    onPressed: () => _copyToClipboard(messages[index].content),
                                    icon: const Icon(Icons.copy, color: Color(0xFFe6af2c)),
                                    tooltip: 'Copy to Clipboard',
                                  ),
                                  IconButton(
                                    onPressed: () => _deleteMessage(index),
                                    icon: const Icon(Icons.delete, color: Colors.red),
                                    tooltip: 'Delete',
                                  ),
                                ],
                              ),
                            ],
                          ),
                          children: [
                            Padding(
                              padding: const EdgeInsets.all(16.0),
                              child: Text(
                                messages[index].content,
                                style: const TextStyle(color: Colors.white),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class MessageData {
  final String id;
  final String content;

  MessageData({required this.id, required this.content});
}
