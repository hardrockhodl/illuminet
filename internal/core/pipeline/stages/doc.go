// Package stages provides reusable [pipeline.Stage] implementations
// for enriching a [model.Sample] between adapter receipt and exporter
// dispatch.
//
// Currently implemented:
//
//   - [PeerClassificationStage] sets Peer.Type from LLDP capabilities
//     and SystemDescription, encoding the operational rules described
//     in docs/ARCHITECTURE.md including the Docker-on-Linux quirk
//     where Linux hosts advertise bridge or router capability because
//     of container networking.
//   - [PortClassificationStage] sets Interface.Classification from
//     Peer.Type, encoding paregupt/nexus_traffic_monitor's
//     Core/Edge split.
//
// Stages have no external dependencies; they are pure transformations
// suitable for unit testing in isolation.
package stages
