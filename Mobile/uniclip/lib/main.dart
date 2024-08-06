import 'dart:io';

import 'package:flutter/material.dart';
import 'package:fluttertoast/fluttertoast.dart';
import 'package:clipboard/clipboard.dart';

void main() {
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'UniClip Client',
      theme: ThemeData(
        primarySwatch: Colors.blue,
      ),
      home: const MyHomePage(title: 'UniClip Client'),
    );
  }
}

class MyHomePage extends StatefulWidget {
  const MyHomePage({super.key, required this.title});

  final String title;

  @override
  _MyHomePageState createState() => _MyHomePageState();
}

class _MyHomePageState extends State<MyHomePage> {
  final TextEditingController _ipController = TextEditingController();
  late WebSocket _webSocket;
  bool _isConnected = false;

  void _connectToServer() async {
    final String ip = _ipController.text;
    try {
      _webSocket = await WebSocket.connect('ws://$ip');
      setState(() {
        _isConnected = true;
      });
      _webSocket.listen((message) {
        FlutterClipboard.copy(message).then((_) {
          Fluttertoast.showToast(msg: 'Clipboard updated');
        });
      });
      Fluttertoast.showToast(msg: 'Connected to server');
    } catch (e) {
      Fluttertoast.showToast(msg: 'Failed to connect to server: $e');
    }
  }

  void _monitorClipboard() {
    FlutterClipboard.paste().then((value) {
      _webSocket.add(value);
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.title),
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: <Widget>[
            Padding(
              padding: const EdgeInsets.all(8.0),
              child: TextField(
                controller: _ipController,
                decoration: const InputDecoration(
                  border: OutlineInputBorder(),
                  labelText: 'Server IP',
                ),
              ),
            ),
            ElevatedButton(
              onPressed: _isConnected ? null : _connectToServer,
              child: const Text('Connect'),
            ),
            if (_isConnected)
              ElevatedButton(
                onPressed: _monitorClipboard,
                child: const Text('Send Clipboard'),
              ),
          ],
        ),
      ),
    );
  }
}
