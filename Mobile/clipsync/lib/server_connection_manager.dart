import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:fluttertoast/fluttertoast.dart';
import 'dart:io';

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
            (data) {
          final receivedData = String.fromCharCodes(data);
          Clipboard.setData(ClipboardData(text: receivedData));
          Fluttertoast.showToast(msg: "Data copied to clipboard");
        },
        onError: (error) {
          Fluttertoast.showToast(msg: "Error: $error");
        },
        onDone: () {
          _isConnected = false;
          Fluttertoast.showToast(msg: "Disconnected from server");
        },
      );
    } catch (e) {
      Fluttertoast.showToast(msg: "Failed to connect: $e");
    }
  }
}
