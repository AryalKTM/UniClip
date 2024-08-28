import 'package:flutter/material.dart';
import 'saved_address_manager.dart';
import 'server_connection_manager.dart';

class MyHomePage extends StatefulWidget {
  const MyHomePage({super.key, required this.title});

  final String title;

  @override
  State<MyHomePage> createState() => _MyHomePageState();
}

class _MyHomePageState extends State<MyHomePage> {
  final TextEditingController _ipController = TextEditingController();
  final TextEditingController _portController = TextEditingController();
  final TextEditingController _nameController = TextEditingController();
  String? _ipError;
  String? _portError;
  bool _isConnected = false;
  ServerConnectionManager? _connectionManager;
  final SavedAddressManager _addressManager = SavedAddressManager();
  List<SavedAddress> _savedAddresses = [];
  SavedAddress? _editingAddress;

  @override
  void initState() {
    super.initState();
    _connectionManager = ServerConnectionManager(
      ipController: _ipController,
      portController: _portController,
    );
    _addressManager.loadAddresses().then((addresses) {
      setState(() {
        _savedAddresses = addresses;
      });
    });
  }

  void _onConnectPressed() async {
    setState(() {
      _ipError = null;
      _portError = null;
    });

    if (!_validateIP(_ipController.text)) {
      setState(() {
        _ipError = "Invalid IP Address";
      });
      return;
    }

    if (!_validatePort(_portController.text)) {
      setState(() {
        _portError = "Invalid Port Number";
      });
      return;
    }

    await _connectionManager?.connectToServer();
    setState(() {
      _isConnected = _connectionManager?.isConnected ?? false;
    });
  }

  bool _validateIP(String ip) {
    final regex = RegExp(r"^(([0-9]{1,3})\.){3}([0-9]{1,3})$");
    return regex.hasMatch(ip);
  }

  bool _validatePort(String port) {
    final regex = RegExp(r"^[0-9]{1,5}$");
    return regex.hasMatch(port);
  }

  void _onSaveAddress() async {
    final address = SavedAddress(
      name: _nameController.text,
      ip: _ipController.text,
      port: _portController.text,
    );

    await _addressManager.saveAddress(address);
    setState(() {
      if (_editingAddress != null) {
        _savedAddresses.remove(_editingAddress);
        _editingAddress = null;
      }
      _savedAddresses.add(address);
    });

    Navigator.pop(context);
  }

  void _onEditAddress(SavedAddress address) {
    setState(() {
      _editingAddress = address;
      _ipController.text = address.ip;
      _portController.text = address.port;
      _nameController.text = address.name;
    });

    _showAddressDialog(isEdit: true);
  }

  void _onAddressPressed(SavedAddress address) {
    setState(() {
      _ipController.text = address.ip;
      _portController.text = address.port;
      _nameController.text = address.name;
    });
  }

  void _onDeleteAddress(SavedAddress address) async {
    await _addressManager.deleteAddress(address);
    setState(() {
      _savedAddresses.remove(address);
    });
  }

  void _showAddressDialog({bool isEdit = false}) {
    showDialog(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: Text(isEdit ? 'Edit Address' : 'Add Address',
              textAlign: TextAlign.center),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextField(
                controller: _nameController,
                decoration: const InputDecoration(labelText: 'Name'),
              ),
              const SizedBox(height: 16),
              TextField(
                keyboardType: TextInputType.number,
                controller: _ipController,
                decoration: InputDecoration(
                    labelText: 'IP Address', errorText: _ipError),
              ),
              const SizedBox(height: 16),
              TextField(
                keyboardType: TextInputType.number,
                controller: _portController,
                decoration:
                    InputDecoration(labelText: 'Port', errorText: _portError),
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () {
                Navigator.pop(context);
              },
              child: const Text('Cancel'),
            ),
            TextButton(
              onPressed: _onSaveAddress,
              child: const Text('Save'),
            ),
          ],
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(widget.title),
      ),
      body: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          children: [
            Text(
              _isConnected ? "Connected" : "Not Connected",
              style: TextStyle(
                  fontSize: 20,
                  color: _isConnected ? Colors.greenAccent : Colors.redAccent),
            ),
            const SizedBox(height: 16),
            TextField(
              controller: _ipController,
              decoration: InputDecoration(
                labelText: 'IP Address',
                errorText: _ipError,
                suffixIcon: const Icon(
                  Icons.account_tree_rounded,
                  color: Color(0xFFefb32e),
                ),
              ),
            ),
            const SizedBox(height: 16.0),
            TextField(
              controller: _portController,
              decoration: InputDecoration(
                labelText: 'Port',
                errorText: _portError,
                suffixIcon: const Icon(
                  Icons.import_export_rounded,
                  color: Color(0xFFefb32e),
                ),
              ),
            ),
            const SizedBox(height: 16.0),
            ElevatedButton(
              onPressed: _onConnectPressed,
              child: Text(_isConnected ? 'Connected' : 'Connect'),
            ),
            Expanded(
              child: ListView.builder(
                itemCount: _savedAddresses.length,
                itemBuilder: (context, index) {
                  final address = _savedAddresses[index];
                  return Dismissible(
                    key: UniqueKey(),
                    onDismissed: (direction) {
                      _onDeleteAddress(address);
                      ScaffoldMessenger.of(context).showSnackBar(
                        SnackBar(
                          content: Text(
                            '${address.name} deleted',
                            style: const TextStyle(color: Colors.white),
                          ),
                          backgroundColor: Colors.black,
                        ),
                      );
                    },
                    background: Container(color: Colors.red),
                    child: ListTile(
                      leading: const Icon(Icons.computer_rounded),
                      iconColor: const Color(0xFFefb32e),
                      title: Text(address.name),
                      titleTextStyle: const TextStyle(
                        color: Color(0xFFefb32e),
                      ),
                      subtitle: Text('${address.ip}:${address.port}'),
                      onTap: () => _onAddressPressed(address),
                      onLongPress: () => _onEditAddress(address),
                    ),
                  );
                },
              ),
            ),
            const SizedBox(height: 16.0),
          ],
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => _showAddressDialog(),
        child: const Icon(Icons.add),
      ),
    );
  }
}
