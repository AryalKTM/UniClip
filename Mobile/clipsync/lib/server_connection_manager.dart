import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:fluttertoast/fluttertoast.dart';
import 'package:clipboard_watcher/clipboard_watcher.dart'; // Ensure correct import
import 'dart:io';
import 'dart:convert'; // Import for utf8 encoding

class ServerConnectionManager {
  final TextEditingController ipController;
  final TextEditingController portController;
  Socket? _socket;
  bool _isConnected = false;

  ServerConnectionManager({
    required this.ipController,
    required this.portController,
  });

  bool get isConnected => _isConnected;

  Future<void> connectToServer() async {
    final ip = ipController.text;
    final port = portController.text;
    String address = '$ip:$port';

    try {
      Fluttertoast.showToast(
          msg: "Connecting to Server",
          toastLength: Toast.LENGTH_SHORT,
          gravity: ToastGravity.CENTER);

      _socket = await Socket.connect(ip, int.parse(port));
      _isConnected = true;
      Fluttertoast.showToast(
          msg: "Connected to Server at: $address",
          toastLength: Toast.LENGTH_SHORT,
          gravity: ToastGravity.CENTER);

      _socket!.listen(
        (data) => _handleReceivedData(data),
        onError: (error) => _handleError(error),
        onDone: _handleDone,
      );

      // Start monitoring the clipboard
      _startMonitoringClipboard();
    } catch (e) {
      Fluttertoast.showToast(msg: "Failed to connect: $e");
    }
  }

  Future<void> sendClipboardData() async {
    ClipboardData? clipboardData = await _getLocalClipboard();

    if (clipboardData != null && clipboardData.text != null) {
      _socket!.add(utf8.encode(clipboardData.text!)); // Send the text as UTF-8 encoded bytes
      Fluttertoast.showToast(msg: "Clipboard data sent to server");
    } else {
      Fluttertoast.showToast(msg: "No text found in clipboard");
    }
  }

  void _handleReceivedData(List<int> data) {
    final receivedData = String.fromCharCodes(data);
    _copyDataToClipboard(receivedData);
    Fluttertoast.showToast(msg: "Data copied to clipboard");
  }

  void _handleError(Object error) {
    Fluttertoast.showToast(msg: "Error: $error");
  }

  void _handleDone() {
    _isConnected = false;
    Fluttertoast.showToast(msg: "Disconnected from server");
    _stopMonitoringClipboard(); // Stop monitoring when disconnected
  }

  void _copyDataToClipboard(String data) {
    Clipboard.setData(ClipboardData(text: data));
  }

  Future<ClipboardData?> _getLocalClipboard() async {
    ClipboardData? localData = await Clipboard.getData("text/plain");
    return localData;
  }

  void _startMonitoringClipboard() {
    ClipboardWatcher.instance.addListener(_onClipboardChanged as ClipboardListener);
    ClipboardWatcher.instance.start();
  }

  void _stopMonitoringClipboard() {
    ClipboardWatcher.instance.removeListener(_onClipboardChanged as ClipboardListener);
    ClipboardWatcher.instance.stop();
  }

  void _onClipboardChanged(ClipboardData? clipboardData) async {
    if (_isConnected && _socket != null && clipboardData?.text != null) {
      _socket!.add(utf8.encode(clipboardData!.text!)); // Send the text as UTF-8 encoded bytes
      Fluttertoast.showToast(msg: "Clipboard data sent to server");
    }
  }
}