import SwiftUI

struct ContentView: View {
    @StateObject private var clipboardMonitor = ClipboardMonitor()
    
    var body: some View {
        VStack {
            Text("Clipboard Monitor")
                .font(.largeTitle)
                .padding()
            
            Button(action: {
                clipboardMonitor.toggleMonitoring()
            }) {
                Text(clipboardMonitor.isMonitoring ? "Stop Monitoring" : "Start Monitoring")
                    .padding()
                    .background(Color.blue)
                    .foregroundColor(.white)
                    .cornerRadius(8)
            }
            .padding()
        }
    }
}

struct ContentView_Previews: PreviewProvider {
    static var previews: some View {
        ContentView()
    }
}
